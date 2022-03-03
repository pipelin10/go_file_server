package main

import (
	"fmt"
	"io"
	"log"
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
		numberBytesRead, err := file.ReadAt(fileBuffer, int64(currentByte))

		currentByte += bufferSize //We move the current byte

		(*clientConnection).Write(fileBuffer[:numberBytesRead]) //We send to the client the package read
		if err == io.EOF {
			fmt.Printf("Complete sending file to %s\n", (*clientConnection).RemoteAddr().String())
			break
		}
	}
	file.Close()
}

//getDataFromClient allows to get data sent from a client
func getDataFromClient(filePath string, dirPath string, clientConnection *net.Conn, bufferSize uint32) {
	//We need a current byte te read through the file
	var currentByte uint32 = 0

	fileBuffer := make([]byte, bufferSize)

	file, err := os.Create(strings.TrimSpace(filePath))
	if err != nil {
		//If the file couldn't be created then a error arise
		fmt.Println("Error creating file")
		if os.IsNotExist(err) {
			//If the directory doesn't exist we need to create it.
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

	for {
		//We read a package of size equals to bufferSize
		numberBytesRead, err := (*clientConnection).Read(fileBuffer)
		fileBufferString := string(fileBuffer[:])

		//If a error arise during buffer read then we break
		if err == io.EOF {
			message := "Complete send file"
			fmt.Println(message)
			break
		}

		if err != nil {
			message := "Error sending file"
			fmt.Println(message)
			break
		}

		//If the file doesn't exist we arise an error
		if fileBufferString == "No se encuentra un archivo en la ruta" {
			fmt.Println("Error al cargar el archivo")
			return
		}

		//We write in the file al the data read until the numberBytesRead
		_, err = file.WriteAt(fileBuffer[:numberBytesRead], int64(currentByte))

		if err == io.EOF {
			fmt.Println(err)
			break
		}

		//If we read all data from the file sent we need to stop
		if uint32(numberBytesRead) != bufferSize {
			message := "Complete send file"
			fmt.Println(message)
			break
		}

		currentByte += bufferSize //We move the current byte
	}
	file.Close()
}
