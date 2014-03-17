// main
package main

import ()

func main() {
	s := NewServer(":443", 8192, "cert.pem", "key.pem", "")
	s.Start()
}
