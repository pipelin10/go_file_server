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

func get_data_from_server(filePath string, dirPath string, conn net.Conn) {
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
		// fmt.Println("File Buffer String", fileBufferString)
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

func send_data_to_server(filePath string, conn net.Conn) {
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

func exist_channel(channel string, arrayChannels []string) bool {
	for _, channelArray := range arrayChannels {
		if channelArray == channel {
			return true
		}
	}
	return false
}

func handleDataRecieved(conn net.Conn) {
	fmt.Printf("Listen to data from %s\n", conn.RemoteAddr().String())
	for {
		netData, err := bufio.NewReader(conn).ReadString('\n')
		if err != nil {
			fmt.Print(err)
			return
		}
		fmt.Print("Recieving Data\n")
		fmt.Print(netData)
		splitMessage := strings.Split(string(netData), " ")
		command := strings.TrimSpace(splitMessage[0])
		if command == "send" {
			fileName := splitMessage[1]
			localAddress := conn.LocalAddr().String()
			localAddress = strings.Replace(localAddress, ".", "_", -1)
			localAddress = strings.Replace(localAddress, ":", "_", -1)
			filePath := ".\\files_recieved_client" + localAddress + "\\" + fileName
			dirPath := ".\\files_recieved_client" + localAddress + "\\"
			get_data_from_server(filePath, dirPath, conn)
		}
		fmt.Print("\n")
	}
}

func main() {
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
		text_split := strings.Split(strings.TrimSpace(string(text)), " ")
		command := strings.TrimSpace(string(text_split[0]))
		if command == "st" {
			fmt.Fprintf(c, text+"\n")
			fmt.Println("TCP Cliente exting")
			c.Close()
			return
		} else if command == "get" {
			fmt.Fprintf(c, text+"\n")
			filePath := ".\\files_recieved_client\\" + text_split[1]
			dirPath := ".\\files_recieved_client\\"
			get_data_from_server(filePath, dirPath, c)
		} else if command == "send" {
			fmt.Fprintf(c, text+"\n")
			filePath := ".\\files_to_send_client\\" + text_split[1]
			send_data_to_server(filePath, c)
		} else if command == "subscribe" {
			channel := text_split[1]
			fmt.Printf(channel)
			if !exist_channel(channel, channelsSubscribed) {
				fmt.Fprintf(c, text+"\n")
				channelsSubscribed = append(channelsSubscribed, channel)
				fmt.Print(channelsSubscribed)
				fmt.Printf("\n")
			} else {
				fmt.Printf("Already subscribed to channel %s\n", channel)
			}
		} else {
			fmt.Fprintf(c, text)
		}
	}
}
