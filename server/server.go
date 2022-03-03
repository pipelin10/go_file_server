package main

import (
	"log"
	"net"
)

var channels map[string][]net.Conn   // var channels corresponds to clients subscribed to a channel
var connChannels map[string][]string // var connChannels corresponds to channels related to a client

func main() {
	channels = make(map[string][]net.Conn)
	connChannels = make(map[string][]string)

	//Initializing pool
	tcpPool, err := InitPool()

	if err != nil {
		log.Fatal(err)
	}

	//Handle connections to the server
	handleConnections(tcpPool)
}
