package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"os"
	"strings"
)

//existChannel allows to find if a specify channel exists
func existChannel(channel string, arrayChannels []string) bool {
	//We look to each channel and see if a coincidence exists between channels
	for _, channelSubscribed := range arrayChannels {
		if channelSubscribed == channel {
			return true
		}
	}
	return false
}

//handleConnection allows to manage the data recieved from the server
func handleDataRecieved(serverConnection *net.Conn) {
	//Const to compare commands
	const send string = "send"

	//Const to paths
	const dirClient string = ".\\files_recieved_client"

	var remoteAddress string = (*serverConnection).RemoteAddr().String()

	fmt.Printf("Listen to data from %s\n", remoteAddress) //Ready to listen data

	for {
		//Get message from server
		serverMessage, err := bufio.NewReader((*serverConnection)).ReadString('\n')

		if err != nil {

			if len(serverMessage) != 0 {
				//If something was sent from the server it shows up before end the process
				fmt.Println(serverMessage)
			}

			log.Fatal(err)
		}

		splitMessage := strings.Split(string(serverMessage), " ")
		command := strings.TrimSpace(splitMessage[0])

		if command == send {

			//We setup a localaddress to create a folder with the data recieved from the server
			localAddress := (*serverConnection).LocalAddr().String()
			localAddress = strings.Replace(localAddress, ".", "_", -1)
			localAddress = strings.Replace(localAddress, ":", "_", -1)

			//We setup the filename and the dirpath needed to read the file and create the folder
			//if it doesn't exist
			fileName := splitMessage[1]
			filePath := dirClient + localAddress + "\\" + fileName
			dirPath := dirClient + localAddress + "\\"
			getDataFromServer(filePath, dirPath, serverConnection)
		}
		fmt.Printf("\n")
	}
}

func handleMessages(serverConnection *net.Conn) {
	//Const to compare commands
	const stop string = "st"
	const send string = "send"
	const subscribe string = "subscribe"

	for {
		message, err := bufio.NewReader(os.Stdin).ReadString('\n') //Read message from the client

		if err != nil {
			log.Fatal(err)
		}

		messageSplit := strings.Split(strings.TrimSpace(string(message)), " ")
		command := strings.TrimSpace(string(messageSplit[0]))

		if command == stop { //Stop connection with server

			fmt.Fprintf((*serverConnection), message+"\n")
			fmt.Println("TCP Cliente exit")

			(*serverConnection).Close()

			return

		} else if command == send {

			fmt.Fprintf((*serverConnection), message+"\n")

			//Check if filename or channel were not specified
			if len(messageSplit) == 1 {
				fmt.Printf("Please specify a filename\n")
				continue
			}
			if len(messageSplit) == 2 {
				fmt.Printf("Please specify a channel\n")
				continue
			}

			//Set up filename and filepath
			fileName := messageSplit[1]
			filePath := ".\\files_to_send_client\\" + fileName

			sendDataToServer(filePath, serverConnection)

		} else if command == subscribe {

			//Check if channel was not specified
			if len(messageSplit) == 1 {
				fmt.Printf("Please specify a channel\n")
				continue
			}

			channel := messageSplit[1]

			//Check if the client was not subscribed to this channel already
			if !existChannel(channel, channelsSubscribed) {

				fmt.Fprintf((*serverConnection), message+"\n") //Subscribe in server

				channelsSubscribed = append(channelsSubscribed, channel)

				fmt.Print(channelsSubscribed, "\n")
			} else {
				fmt.Printf("Already subscribed to channel %s\n", channel)
			}
		} else {
			fmt.Printf("Please specify a valid command (st, send, subscribe)\n")
		}
	}
}
