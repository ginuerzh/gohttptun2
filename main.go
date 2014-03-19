// main
package main

import (
	"flag"
	"log"
)

type Config struct {
	Addr       string
	Server     string
	Proxy      string
	CertFile   string
	KeyFile    string
	IsClient   bool
	BufferSize int
}

var config Config

func init() {
	flag.StringVar(&config.Proxy, "P", "", "proxy for forward")
	flag.StringVar(&config.Server, "S", ":443", "the server that client connecting to")
	flag.StringVar(&config.Addr, "L", ":8888", "listen address")
	flag.StringVar(&config.CertFile, "cert", "cert.pem", "cert.pem file")
	flag.StringVar(&config.KeyFile, "key", "key.pem", "key.pem file")
	flag.BoolVar(&config.IsClient, "c", false, "client")
	flag.IntVar(&config.BufferSize, "b", 4096, "buffer size")
	flag.Parse()

	log.SetFlags(log.LstdFlags | log.Lshortfile)
}

func main() {
	if !config.IsClient {
		s := NewServer(config.Addr, config.CertFile, config.KeyFile, config.Proxy)
		s.Start()
		return
	}

	c := NewClient(config.Addr, config.Server, config.Proxy)
	c.Start()
}
