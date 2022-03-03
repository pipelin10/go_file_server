package main

import (
	"flag"
	"fmt"
	"net"
)

var channelsSubscribed []string //Var to know what channels are client subscribed

func main() {

	//Config for host and port
	host := *flag.String("host", "localhost", "Host to connect")
	port := *flag.Int("port", 8080, "Port to connect")

	address := fmt.Sprintf("%s:%d", host, port)

	serverConnection, err := net.Dial("tcp4", address) //Try to connect to server
	if err != nil {
		fmt.Println(err)
		return
	}
	defer serverConnection.Close()

	go handleDataRecieved(&serverConnection) //Handle all data recieved from server

	handleMessages(&serverConnection) //Handle all messages from the client

}
