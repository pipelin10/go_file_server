package main

import (
	"bufio"
	"fmt"
	"net"
	"strings"
)

func eraseConnChannels(c net.Conn) {
	remoteAddress := c.RemoteAddr().String()
	fmt.Println(connChannels[remoteAddress])
	for _, channel := range connChannels[remoteAddress] {
		for index, conn := range channels[channel] {
			connAddress := conn.RemoteAddr().String()
			if connAddress == remoteAddress {
				channels[channel] = append(channels[channel][:index], channels[channel][index+1:]...)
				break
			}
		}
	}
}

func handleConnection(c net.Conn) {
	const STOP string = "STOP"
	const GET string = "get"
	const SEND string = "send"
	const SUBSCRIBE string = "subscribe"
	const bufferSize uint32 = 1024

	fmt.Printf("Serving %s\n", c.RemoteAddr().String())
	defer c.Close()
	for {
		netData, err := bufio.NewReader(c).ReadString('\n')
		if err != nil {
			fmt.Println(err)
			eraseConnChannels(c)
			return
		}
		splitMessage := strings.Split(string(netData), " ")
		command := strings.TrimSpace(splitMessage[0])
		fmt.Println(splitMessage)
		if command == STOP {
			fmt.Printf("Closing connection with %s\n", c.RemoteAddr().String())
			eraseConnChannels(c)
			break
		} else if command == GET {
			if len(splitMessage) == 1 {
				fmt.Fprintf(c, "Please specify a filename\n")
				continue
			}
			fileName := splitMessage[1]
			filePath := ".\\files_to_send_server\\" + fileName
			sendDataToClient(filePath, c, bufferSize)
		} else if command == SEND {
			if len(splitMessage) == 1 {
				fmt.Fprintf(c, "Please specify a filename\n")
				continue
			} else if len(splitMessage) == 2 {
				fmt.Fprintf(c, "Please specify a channel\n")
				continue
			}
			fileName := splitMessage[1]
			channel := splitMessage[2]
			filePath := ".\\files_recieved_server\\" + fileName
			getDataFromClient(filePath, c, bufferSize)
			fmt.Print(channel)
			ipHostClientSending := c.RemoteAddr().String()
			for _, conn := range channels[channel] {
				ipHostCienteReceiving := conn.RemoteAddr().String()
				if ipHostClientSending != ipHostCienteReceiving {
					fmt.Fprintf(conn, "send %s\n", fileName)
					sendDataToClient(filePath, conn, bufferSize)
				}
			}
		} else if command == SUBSCRIBE {
			if len(splitMessage) == 1 {
				fmt.Fprintf(c, "Please specify a channel\n")
				continue
			}
			channel := splitMessage[1]
			channels[channel] = append(channels[channel], c)
			connChannels[c.RemoteAddr().String()] = append(connChannels[c.RemoteAddr().String()], channel)
			for chanMap, connArray := range channels {
				fmt.Print(chanMap)
				for conn := range connArray {
					fmt.Print(" ", connArray[conn].RemoteAddr().String())
				}
				fmt.Printf("\n")
			}
		} else {
			fmt.Fprintf(c, "Please specify a valid command\n")
		}
	}
	c.Close()
}

func handleConnections(tcpPool *TcpConnPool) {
	address := fmt.Sprintf("%s:%d", tcpPool.Host, tcpPool.Port)

	listener, err := net.Listen("tcp4", address)

	if err != nil {
		fmt.Println(err)
		return
	}
	defer listener.Close()

	fmt.Println("Now listen on address", address)

	for {

		connectionClient, err := listener.Accept()

		if err != nil {
			fmt.Println(err)
			return
		}
		defer connectionClient.Close()

		if tcpPool.NumOpen >= tcpPool.MaxOpenConn {
			fmt.Fprintf(connectionClient, "Can't establish a connection with server\n")
			connectionClient.Close()
			continue
		}

		tcpPool.Mu.Lock()

		tcpPool.NumOpen++

		tcpConnClient := &TcpConn{
			Id:   address,
			Pool: tcpPool,
			Conn: connectionClient,
		}
		tcpPool.Connections = append(tcpPool.Connections, tcpConnClient)

		tcpPool.Mu.Unlock()

		go handleConnection(connectionClient)
	}
}
