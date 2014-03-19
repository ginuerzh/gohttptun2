// server
package main

import (
	"bufio"
	"bytes"
	"io/ioutil"
	"log"
	"net"
	"net/http"
)

type Server struct {
	Addr     string
	Proxy    string
	CertFile string
	KeyFile  string
}

func NewServer(addr string, certFile string, keyFile string, proxy string) *Server {
	return &Server{
		Addr:     addr,
		CertFile: certFile,
		KeyFile:  keyFile,
		Proxy:    proxy,
	}
}

func (s *Server) Start() {
	log.Fatal(listenAndServeTLS(s.Addr, s.CertFile, s.KeyFile, s.handleConnection))
}

func (this *Server) handleConnection(c net.Conn) {
	req, err := http.ReadRequest(bufio.NewReader(c))
	if err != nil {
		log.Println(err)
		return
	}

	// secure connection
	if req.Method == "CONNECT" {
		s, status, err := connect(req.URL.Host, this.Proxy, false)
		//log.Println("CONNECT", req.URL.Host, status)
		if s != nil {
			defer s.Close()
		}
		if len(status) > 0 {
			//log.Println(status)
			w := bufio.NewWriter(c)
			w.WriteString(status)
			err = w.Flush()
		}
		if err != nil {
			log.Println(err)
			return
		}

		transfer(s, c)
		return
	}

	resp, err := this.doRequest(req)
	if resp != nil {
		defer resp.Body.Close()
	}
	if err != nil {
		log.Println(err)
		return
	}

	resp.Write(c)
}

func (this *Server) doRequest(req *http.Request) (resp *http.Response, err error) {
	if len(this.Proxy) > 0 {
		proxy, err := net.Dial("tcp", this.Proxy)
		if err != nil {
			return nil, err
		}
		defer proxy.Close()

		if err := req.WriteProxy(proxy); err != nil {
			log.Println(err)
			return nil, err
		}

		r, err := ioutil.ReadAll(proxy)
		if err != nil {
			log.Println(err)
			return nil, err
		}
		return http.ReadResponse(bufio.NewReader(bytes.NewBuffer(r)), req)
	}

	req.Header.Del("Proxy-Connection")
	req.RequestURI = ""
	return http.DefaultClient.Do(req)
}
