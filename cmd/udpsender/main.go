package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"os"
)

const network = "udp"
const address = "localhost:42069"

func main() {
	udpAddr, err := net.ResolveUDPAddr(network, address)
	if err != nil {
		log.Fatalf("Failed to resolve udp address on port %s: %s\n", address, err)
	}
	conn, err := net.DialUDP(network, nil, udpAddr)
	if err != nil {
		log.Fatalf("Failed to dial udp on address %s: %s\n", udpAddr.String(), err)
	}
	defer conn.Close()
	buf := bufio.NewReader(os.Stdin)
	for {
		fmt.Printf("> ")
		data, err := buf.ReadBytes('\n')
		if err != nil {
			log.Println(err)
			continue
		}
		if _, err = conn.Write(data); err != nil {
			log.Println(err)
		}
	}
}
