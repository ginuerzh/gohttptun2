// client
package main

import (
	"log"
	"net"
)

type Client struct {
	Addr   string
	Server string
	Proxy  string
}

func NewClient(addr string, server string, proxy string) *Client {
	return &Client{
		Addr:   addr,
		Server: server,
		Proxy:  proxy,
	}
}

func (c *Client) Start() {
	log.Fatal(listenAndServe(c.Addr, c.handleConnection))
}

func (c *Client) handleConnection(conn net.Conn) {
	s, err := connect(c.Server, c.Proxy, true)
	if err != nil {
		log.Println(err)
		return
	}
	defer s.Close()

	transfer(conn, s)
}
