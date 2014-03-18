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
		s, err := connect(req.URL.Host, this.Proxy, false)
		if err != nil {
			log.Println(err)
			return
		}
		defer s.Close()

		w := bufio.NewWriter(c)
		w.WriteString("HTTP/1.1 200 Connection established\r\n")
		w.WriteString("Proxy-agent: gost/1.0\r\n\r\n")
		if err := w.Flush(); err != nil {
			log.Println(err)
			return
		}

		transfer(c, s)
	}
}
