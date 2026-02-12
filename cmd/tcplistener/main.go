package main

import (
	"fmt"
	"log"
	"net"
	"httpfromtcp/internal/request"
)

const (
	port = ":42069"
)

func main() {
	listener, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatalf("Error listening on port '%s': %s", port, err)
	}
	defer listener.Close()
	fmt.Println("Listening for TCP traffic on", port)
	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Fatalf("couldn't accept connection: %s\n", err)
		}
		fmt.Println("Accepted connection from", conn.RemoteAddr())
		req, err := request.RequestFromReader(conn)
		if err != nil {
			log.Fatalf("couldn't parse request: %s\n", err)
		}
		fmt.Println("Request line:")
		fmt.Println(" - Method:", req.RequestLine.Method)
		fmt.Println(" - Target:", req.RequestLine.RequestTarget)
		fmt.Println(" - Version:", req.RequestLine.HttpVersion)
		fmt.Println("Headers:")
		for key, value := range req.Headers {
			fmt.Println(" -", key + ":", value)
		}
		fmt.Println("Body:")
		fmt.Println(string(req.Body))
		conn.Close()
		fmt.Println("Connection to", conn.RemoteAddr(), "closed")
	}
}
