// util
package main

import (
	"bufio"
	"crypto/tls"
	"errors"
	"io"
	"log"
	"net"
	"net/http"
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

func connect(addr string, proxy string, secure bool) (conn net.Conn, err error) {
	//log.Println("CONNECT", addr, "proxy", proxy, "secure", secure)
	if len(proxy) > 0 {
		conn, err = net.Dial("tcp", proxy)
		if err != nil {
			return
		}
		w := bufio.NewWriter(conn)
		w.WriteString("CONNECT " + addr + " HTTP/1.1\r\n")
		w.WriteString("Host: " + addr + "\r\n")
		w.WriteString("Proxy-Connection: keep-alive\r\n\r\n")

		if err := w.Flush(); err != nil {
			return nil, err
		}
		//log.Println("write ok")
		resp, err := http.ReadResponse(bufio.NewReader(conn), nil)
		if err != nil {
			conn.Close()
			return nil, err
		}
		defer resp.Body.Close()

		//log.Println(resp.Status)
		if resp.StatusCode != http.StatusOK {
			conn.Close()
			return nil, errors.New(resp.Status)
		}
	} else {
		conn, err = net.Dial("tcp", addr)
		if err != nil {
			return
		}
	}

	if secure {
		config := &tls.Config{InsecureSkipVerify: true}
		cli := tls.Client(conn, config)
		if err := cli.Handshake(); err != nil {
			cli.Close()
			return nil, err
		}
		//log.Println("hand shake done")
		conn = cli
	}

	//log.Println("CONNECT", addr, "OK")
	return conn, nil
}

func transfer(src, dst net.Conn) {
	//log.Println("start transfer")
	statusChan := make(chan int, 1)

	go func(r io.Reader, w io.Writer) {
		for {
			/*
				b := make([]byte, 1460)
				n, err := r.Read(b)
				if err != nil {
					log.Println(err)
					break
				}
				log.Println("r", n)
				_, err = w.Write(b[:n])
				if err != nil {
					log.Println(err)
					break
				}
			*/
			_, err := io.Copy(w, r)
			if err != nil {
				log.Println(err)
				break
			}
		}

		statusChan <- 1
	}(src, dst)

	for {
		/*
			b := make([]byte, 1460)
			n, err := dst.Read(b)
			if err != nil {
				log.Println(err)
				break
			}
			log.Println("r", n)
			_, err = src.Write(b[:n])
			if err != nil {
				log.Println(err)
				break
			}
		*/
		_, err := io.Copy(src, dst)
		if err != nil {
			log.Println(err)
			break
		}
	}

	<-statusChan
}
