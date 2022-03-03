package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"strings"
)

//eraseConnChannels allows to erase the client from the channels related to the client
//also it allows to erase client's channels
func eraseConnChannels(clientConnection *net.Conn) error {
	remoteAddress := (*clientConnection).RemoteAddr().String() //Client's address

	for _, channel := range connChannels[remoteAddress] {

		for index, conn := range channels[channel] {

			connAddress := conn.RemoteAddr().String()
			if connAddress == remoteAddress {
				//Delete client form channel
				channels[channel] = append(channels[channel][:index], channels[channel][index+1:]...)
				break
			}

		}

	}

	delete(connChannels, remoteAddress) //Delete client's channels

	return nil
}

//eraseConnPool allows to erase a connection from the pool and close connection
func closeTcpConn(clientConnection *net.Conn, tcpPool *TcpConnPool) error {
	for index, connPool := range tcpPool.Connections {
		if connPool == clientConnection {
			tcpPool.Connections = append(tcpPool.Connections[:index], tcpPool.Connections[index+1:]...)
		}
	}

	(*clientConnection).Close()

	return nil
}

//handleConnection allows to manage the messages recieved from the client
func handleConnection(clientConnection *net.Conn, tcpPool *TcpConnPool) {
	//Const to compare commands
	const STOP string = "STOP"
	const STOPCL string = "st"
	const SEND string = "send"
	const SUBSCRIBE string = "subscribe"

	const bufferSize uint32 = 1024

	var remoteAddress string = (*clientConnection).RemoteAddr().String()

	//Const showing paths
	const PATH_FILES_SERVER string = ".\\files_server\\"
	var PATH_CLIENT string = fmt.Sprintf("%s\\", remoteAddress)
	PATH_CLIENT = strings.Replace(PATH_CLIENT, ".", "_", -1)
	PATH_CLIENT = strings.Replace(PATH_CLIENT, ":", "_", -1)

	fmt.Printf("Serving %s\n", remoteAddress)

	defer closeTcpConn(clientConnection, tcpPool)

	for {
		serverMessage, err := bufio.NewReader((*clientConnection)).ReadString('\n')

		//If something fails during a message read from client it needs to be managed
		if err != nil {

			fmt.Println(err)

			tcpPool.Mu.Lock()
			defer tcpPool.Mu.Unlock()

			//Erasing channels
			errEraseChannel := eraseConnChannels(clientConnection)
			if errEraseChannel != nil {
				log.Fatal(errEraseChannel)
			}
			//Closing connection
			errEraseConnPool := closeTcpConn(clientConnection, tcpPool)
			if errEraseConnPool != nil {
				log.Fatal(err)
			}

			return
		}

		splitMessage := strings.Split(string(serverMessage), " ")
		command := strings.TrimSpace(splitMessage[0])

		if command == STOP {
			fmt.Printf("Closing connection with %s\n", remoteAddress)

			tcpPool.Mu.Lock()
			defer tcpPool.Mu.Unlock()

			eraseConnChannels(clientConnection)
			closeTcpConn(clientConnection, tcpPool)

			break
		} else if command == SEND {
			//Check if filename or channel were not specify
			if len(splitMessage) == 1 {
				fmt.Fprintf((*clientConnection), "Please specify a filename\n")
				continue
			} else if len(splitMessage) == 2 {
				fmt.Fprintf((*clientConnection), "Please specify a channel\n")
				continue
			}

			//Creating path to file
			fileName := splitMessage[1]
			filePath := PATH_FILES_SERVER + PATH_CLIENT + fileName
			dirPath := PATH_FILES_SERVER + PATH_CLIENT

			channel := splitMessage[2]

			getDataFromClient(filePath, dirPath, clientConnection, bufferSize)

			//We need to send all data to clients subscribed to a channel
			//we look at each client in a channel and send data to this client
			ipHostClientSending := remoteAddress
			for _, conn := range channels[channel] {

				ipHostCienteReceiving := conn.RemoteAddr().String()

				//If client isn't the same as sender then we need to send data
				if ipHostClientSending != ipHostCienteReceiving {
					fmt.Fprintf(conn, "send %s\n", fileName)
					sendDataToClient(filePath, &conn, bufferSize)
				}

			}

		} else if command == SUBSCRIBE {
			//Check if channel was not specify
			if len(splitMessage) == 1 {
				fmt.Fprintf((*clientConnection), "Please specify a channel\n")
				continue
			}
			channel := splitMessage[1]

			//Lock to prevent race condition
			tcpPool.Mu.Lock()

			//Checking channels
			channels[channel] = append(channels[channel], (*clientConnection))
			connChannels[remoteAddress] = append(connChannels[remoteAddress], channel)

			tcpPool.Mu.Unlock()

		} else { //No valid command
			fmt.Fprintf((*clientConnection), "Please specify a valid command\n")
			fmt.Println(tcpPool.Connections)
		}
	}
	//Closing connection
	closeTcpConn(clientConnection, tcpPool)
}

//handleConnections allows to open a new connection and establish if it can be opened. If
//the connection can't be established then a message is and the connection will be closed.
func handleConnections(tcpPool *TcpConnPool) {
	address := fmt.Sprintf("%s:%d", tcpPool.Host, tcpPool.Port)

	listener, err := net.Listen("tcp4", address) // Listen to the address load in config

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

		//If the connection can't be established the server will close that connection
		if tcpPool.NumOpen >= tcpPool.MaxOpenConn {
			fmt.Fprintf(connectionClient, "Can't establish a connection with server\n")
			connectionClient.Close()
			continue
		}

		//New connection establish so it needs to be counted for the pool
		tcpPool.Mu.Lock()

		tcpPool.NumOpen++
		tcpPool.Connections = append(tcpPool.Connections, &connectionClient)

		tcpPool.Mu.Unlock()

		go handleConnection(&connectionClient, tcpPool)
	}
}
