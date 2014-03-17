// server
package main

import (
	"crypto/tls"
	"io"
	"log"
	"net"
)

type Server struct {
	Addr       string
	Proxy      string
	CertFile   string
	KeyFile    string
	BufferSize int
}

func NewServer(addr string, bufSize int, certFile string, keyFile string, proxy string) *Server {
	return &Server{
		Addr:       addr,
		CertFile:   certFile,
		KeyFile:    keyFile,
		BufferSize: bufSize,
		Proxy:      proxy,
	}
}

func (s *Server) Start() {
	cert, err := tls.LoadX509KeyPair(s.CertFile, s.KeyFile)
	if err != nil {
		log.Fatal(err)
	}
	config := &tls.Config{Certificates: []tls.Certificate{cert}}
	l, err := tls.Listen("tcp", s.Addr, config)
	if err != nil {
		log.Fatal(err)
	}
	defer l.Close()

	for {
		conn, err := l.Accept()
		if err != nil {
			log.Fatal(err)
		}

		go func(c net.Conn) {
			io.Copy(c, c)
		}(conn)
	}
}
