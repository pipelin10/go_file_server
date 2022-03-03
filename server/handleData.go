package main

import (
	"fmt"
	"io"
	"net"
	"os"
	"strings"
)

//sendDataToClient allows to send data from server to a client
func sendDataToClient(filePath string, clientConnection *net.Conn, bufferSize uint32) {
	//We need a current byte te read through the file
	var currentByte uint32 = 0

	fileBuffer := make([]byte, bufferSize)

	file, err := os.Open(strings.TrimSpace(filePath))
	if err != nil {
		fmt.Fprintf((*clientConnection), "No se encuentra un archivo en la ruta")
		return
	}
	defer file.Close()

	for {
		//We read from the file
		n, err := file.ReadAt(fileBuffer, int64(currentByte))

		currentByte += bufferSize //We move the current byte

		(*clientConnection).Write(fileBuffer[:n]) //We send to the client the package read
		if err == io.EOF {
			fmt.Fprintf((*clientConnection), "Error sending file\n")
			break
		}
	}
	file.Close()
}

//getDataFromClient allows to get data sent from a client
func getDataFromClient(filePath string, clientConnection *net.Conn, bufferSize uint32) {
	//We need a current byte te read through the file
	var currentByte uint32 = 0

	fileBuffer := make([]byte, bufferSize)

	file, err := os.Create(strings.TrimSpace(filePath))
	if err != nil {
		//If the file couldn't be created then a error arise
		fmt.Println("Error creating file")
		return
	}
	defer file.Close()

	for {
		//We read a package of size equals to bufferSize
		n, err := (*clientConnection).Read(fileBuffer)
		fileBufferString := string(fileBuffer[:])

		//If a error arise during buffer read then we break
		if err == io.EOF || err != nil {
			message := "Error sending file"
			fmt.Println(message)
			fmt.Fprintln((*clientConnection), message)
			break
		}

		if fileBufferString == "No se encuentra un archivo en la ruta" {
			fmt.Println("Error al cargar el archivo")
			return
		}

		//We write
		_, err = file.WriteAt(fileBuffer[:n], int64(currentByte))

		if err == io.EOF {
			fmt.Println(err)
			break
		}

		//If we read all data from the file sent we need to stop
		if uint32(n) != bufferSize {
			break
		}

		currentByte += bufferSize //We move the current byte
	}
	file.Close()
}
