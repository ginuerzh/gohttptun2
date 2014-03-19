// util
package main

import (
	"bufio"
	"crypto/tls"
	"errors"
	"io"
	"log"
	"net"
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

		var b []byte
		b, err = read(conn)
		if err != nil {
			return
		}
		status = string(b)
		if !strings.Contains(status, "200") {
			err = errors.New(status)
			return
		}
		//log.Println(status)
		//log.Println("CONNECT", addr, proxy, "OK")
	} else {
		conn, err = net.Dial("tcp", addr)
		if err != nil {
			return
		}
	}

	if secure {
		config := &tls.Config{InsecureSkipVerify: true}
		cli := tls.Client(conn, config)
		err = cli.Handshake()
		conn = cli
	}

	status = "HTTP/1.0 200 Connection established\r\n" +
		"Proxy-Agent: gost/1.0\r\n\r\n"
	return
}

func transfer(src, dst net.Conn) {
	log.Println("start transfer")
	//statusChan := make(chan int, 1)

	go func(r io.Reader, w io.Writer) {
		readWrite(r, w)
	}(dst, src)

	readWrite(src, dst)

	log.Println("transfer done")
}

func readWrite(r io.Reader, w io.Writer) (err error) {
	var b []byte
	for {
		b, err = read(r)
		if len(b) > 0 {
			if _, err := w.Write(b); err != nil {
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

func read(r io.Reader) ([]byte, error) {
	b := make([]byte, config.BufferSize)
	n, err := r.Read(b)
	return b[:n], err
}
