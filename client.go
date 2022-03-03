package main

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"strings"
)

var channelsSubscribed []string

func getDataFromServer(filePath string, dirPath string, conn net.Conn) {
	const bufferSize = 1024
	currentByte := 0

	fileBuffer := make([]byte, bufferSize)

	file, err := os.Create(strings.TrimSpace(filePath))
	if err != nil {
		fmt.Println("Error creating file")
		if os.IsNotExist(err) {
			err := os.Mkdir(strings.TrimSpace(dirPath), os.ModePerm)
			if err != nil {
				log.Fatal(err)
			}
			file, _ = os.Create(strings.TrimSpace(filePath))
			fmt.Println("Created directory", dirPath)
		} else {
			log.Fatal(err)
		}
	}
	defer file.Close()
	// defer conn.Close()

	for {
		fmt.Println("Inside reading")
		fmt.Println("Before reading")
		n, err := conn.Read(fileBuffer)
		fileBufferString := string(fileBuffer[:])
		if err == io.EOF || err != nil {
			fmt.Println("Algarete!!!")
			break
		}
		if fileBufferString == "No se encuentra un archivo en la ruta" {
			fmt.Println("Error al cargar el archivo")
			return
		}
		fmt.Println("Bytes read", n)
		fmt.Println("File Buffer", fileBuffer)
		bufferFile := bytes.NewBuffer(fileBuffer)
		fmt.Println("File Buffer Lenght: ", bufferFile.Len())
		// cleanedFileBuffer := bytes.Trim(fileBuffer, "\x00")

		fmt.Println("Writing in File")
		_, err = file.WriteAt(fileBuffer[:n], int64(currentByte))

		fmt.Println("Adding buffer")
		currentByte += bufferSize

		if err == io.EOF || n != bufferSize {
			fmt.Println("Algarete!!!")
			break
		}
	}
	// c.Write([]byte("get " + filePath + "\n"))
	fmt.Println("Out reading")
	file.Close()
}

func sendDataToServer(filePath string, conn net.Conn) {
	const bufferSize = 1024
	currentByte := 0
	fileBuffer := make([]byte, bufferSize)

	file, err := os.Open(strings.TrimSpace(filePath))
	fmt.Println("File:", file)
	if err != nil {
		fmt.Fprintf(conn, "No se encuentra un archivo en la ruta")
		return
	}
	defer file.Close()

	for {
		n, err := file.ReadAt(fileBuffer, int64(currentByte))
		currentByte += bufferSize
		fmt.Println(fileBuffer)
		conn.Write(fileBuffer[:n])
		fmt.Println("Sent", n, "bytes")
		if err == io.EOF {
			break
		}
	}
	fmt.Println("Closing File")
	file.Close()
}

func existChannel(channel string, arrayChannels []string) bool {
	for _, channelArray := range arrayChannels {
		if channelArray == channel {
			return true
		}
	}
	return false
}

func handleDataRecieved(conn net.Conn) {
	const SEND string = "send"

	fmt.Printf("Listen to data from %s\n", conn.RemoteAddr().String())
	for {
		netData, err := bufio.NewReader(conn).ReadString('\n')
		if err != nil {
			if len(netData) != 0 {
				fmt.Println(netData)
			}
			log.Fatal(err)
		}
		fmt.Print("Recieving Data\n")
		fmt.Print(netData)
		splitMessage := strings.Split(string(netData), " ")
		command := strings.TrimSpace(splitMessage[0])
		if command == SEND {
			fileName := splitMessage[1]
			localAddress := conn.LocalAddr().String()
			localAddress = strings.Replace(localAddress, ".", "_", -1)
			localAddress = strings.Replace(localAddress, ":", "_", -1)
			filePath := ".\\files_recieved_client" + localAddress + "\\" + fileName
			dirPath := ".\\files_recieved_client" + localAddress + "\\"
			getDataFromServer(filePath, dirPath, conn)
		}
		fmt.Print("\n")
	}
}

func main() {
	const STOP string = "st"
	const GET string = "get"
	const SEND string = "send"
	const SUBSCRIBE string = "subscribe"

	arguments := os.Args
	if len(arguments) == 1 {
		fmt.Println("Please provide a host:port")
		return
	}

	CONNECT := arguments[1]
	c, err := net.Dial("tcp4", CONNECT)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer c.Close()

	go handleDataRecieved(c)

	for {
		reader := bufio.NewReader(os.Stdin)
		text, _ := reader.ReadString('\n')
		textSplit := strings.Split(strings.TrimSpace(string(text)), " ")
		command := strings.TrimSpace(string(textSplit[0]))
		if command == STOP {
			fmt.Fprintf(c, text+"\n")
			fmt.Println("TCP Cliente exit")
			c.Close()
			return
		} else if command == GET {
			fmt.Fprintf(c, text+"\n")
			if len(textSplit) == 1 {
				fmt.Printf("Please specify a filename\n")
				continue
			}
			fileName := textSplit[1]
			filePath := ".\\files_recieved_client\\" + fileName
			dirPath := ".\\files_recieved_client\\"
			getDataFromServer(filePath, dirPath, c)
		} else if command == SEND {
			fmt.Fprintf(c, text+"\n")
			if len(textSplit) == 1 {
				fmt.Printf("Please specify a filename\n")
				continue
			}
			if len(textSplit) == 2 {
				fmt.Printf("Please specify a channel\n")
				continue
			}
			fileName := textSplit[1]
			filePath := ".\\files_to_send_client\\" + fileName
			sendDataToServer(filePath, c)
		} else if command == SUBSCRIBE {
			if len(textSplit) == 1 {
				fmt.Printf("Please specify a channel\n")
				continue
			}
			channel := textSplit[1]
			fmt.Print(channel)
			if !existChannel(channel, channelsSubscribed) {
				fmt.Fprintf(c, text+"\n")
				channelsSubscribed = append(channelsSubscribed, channel)
				fmt.Print(channelsSubscribed)
				fmt.Printf("\n")
			} else {
				fmt.Printf("Already subscribed to channel %s\n", channel)
			}
		} else {
			fmt.Fprint(c, text)
		}
	}
}
