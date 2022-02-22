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

func get_data_from_server(fileName string, conn net.Conn) {
	const bufferSize = 1024
	currentByte := 0

	fileBuffer := make([]byte, bufferSize)

	file, err := os.Create(strings.TrimSpace(fileName))
	if err != nil {
		fmt.Println("Error creating file")
		log.Fatal(err)
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
	// c.Write([]byte("get " + fileName + "\n"))
	fmt.Println("Out reading")
	file.Close()
}

func send_data_to_server(fileName string, conn net.Conn) {
	const bufferSize = 1024
	currentByte := 0
	fileBuffer := make([]byte, bufferSize)

	file, err := os.Open(strings.TrimSpace(fileName))
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

	for {
		reader := bufio.NewReader(os.Stdin)
		fmt.Print(">> ")
		text, _ := reader.ReadString('\n')
		fmt.Fprintf(c, text+"\n")
		text_split := strings.Split(strings.TrimSpace(string(text)), " ")
		if strings.TrimSpace(string(text_split[0])) == "st" {
			fmt.Println("TCP Cliente exting")
			c.Close()
			return
		} else if strings.TrimSpace(string(text_split[0])) == "get" {
			fileName := ".\\files_recieved_client\\" + text_split[1]
			get_data_from_server(fileName, c)
		} else if strings.TrimSpace(string(text_split[0])) == "send" {
			fileName := ".\\files_to_send_client\\" + text_split[1]
			send_data_to_server(fileName, c)
		} else {
			message, err := bufio.NewReader(c).ReadString('\n')
			if err != nil {
				fmt.Println(err)
				fmt.Println("Close Connection")
				return
			}
			fmt.Print("->:" + message)
		}
	}
}
