// util
package main

import (
	"bufio"
	"crypto/tls"
	"errors"
	"io"
	"log"
	"net"
	//"net/http"
	"strings"
)

type HandlerFunc func(net.Conn)

func listenAndServe(addr string, handler HandlerFunc) error {
	l, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}
	defer l.Close()

	for {
		conn, err := l.Accept()
		if err != nil {
			return err
		}

		go func(c net.Conn) {
			defer c.Close()
			handler(c)
		}(conn)
	}

	return nil
}

func listenAndServeTLS(addr string, certFile, keyFile string, handler HandlerFunc) error {
	cert, err := tls.LoadX509KeyPair(certFile, keyFile)
	if err != nil {
		return err
	}

	config := &tls.Config{Certificates: []tls.Certificate{cert}}
	l, err := tls.Listen("tcp", addr, config)
	if err != nil {
		return err
	}
	defer l.Close()

	for {
		conn, err := l.Accept()
		if err != nil {
			return err
		}

		go func(c net.Conn) {
			defer c.Close()
			handler(c)
		}(conn)
	}

	return nil
}

func connect(addr string, proxy string, secure bool) (conn net.Conn, status string, err error) {
	//log.Println("connect", addr, "proxy", proxy, "secure", secure)
	if len(proxy) > 0 {
		conn, err = net.Dial("tcp", proxy)
		if err != nil {
			return
		}
		w := bufio.NewWriter(conn)
		w.WriteString("CONNECT " + addr + " HTTP/1.1\r\n")
		w.WriteString("Host: " + addr + "\r\n")
		w.WriteString("Proxy-Connection: keep-alive\r\n\r\n")
		if err = w.Flush(); err != nil {
			return
		}

		r := bufio.NewReader(conn)
		status, err = r.ReadString('\n')
		s, _ := r.ReadString('\n')
		status += s
		if err != nil {
			return
		}
		if !strings.Contains(status, "200") {
			err = errors.New(status)
			return
		}
		//log.Println("CONNECT", addr, proxy, "OK")
	} else {
		conn, err = net.Dial("tcp", addr)
		if err != nil {
			return
		}
		status = "HTTP/1.1 200 Connection established\r\n\r\n"
	}

	if secure {
		config := &tls.Config{InsecureSkipVerify: true}
		cli := tls.Client(conn, config)
		err = cli.Handshake()
		conn = cli
	}

	return
}

func transfer(src, dst net.Conn) {
	log.Println("start transfer")
	statusChan := make(chan int, 1)

	go func(r io.Reader, w io.Writer) {
		readWrite(r, w)
		statusChan <- 1
	}(src, dst)

	readWrite(dst, src)
	<-statusChan

	log.Println("transfer done")
}

func readWrite(r io.Reader, w io.Writer) (err error) {
	n := 0
	b := make([]byte, config.BufferSize)

	for {
		//log.Println("read...")
		n, err = r.Read(b)
		//log.Println("r", n)
		if n > 0 {
			if _, err := w.Write(b[:n]); err != nil {
				return err
			}
		}
		if err != nil {
			if err != io.EOF {
				log.Println(err)
			}
			break
		}
	}

	return
}
