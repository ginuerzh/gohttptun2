// server
package main

import (
	"bufio"
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
		log.Println("CONNECT", req.URL.Host, status)
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

		transfer(c, s)
	}
}
