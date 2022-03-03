package main

import (
	"log"
	"net"
)

var channels map[string][]net.Conn
var connChannels map[string][]string

func main() {
	channels = make(map[string][]net.Conn)
	connChannels = make(map[string][]string)

	tcpPool, err := InitPool()

	if err != nil {
		log.Fatal(err)
	}

	handleConnections(tcpPool)
}
