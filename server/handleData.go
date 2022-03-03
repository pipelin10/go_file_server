package main

import (
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"strings"
)

func sendDataToClient(filePath string, clientConnection *net.Conn, bufferSize uint32) {
	var currentByte uint32 = 0

	fileBuffer := make([]byte, bufferSize)

	file, err := os.Open(strings.TrimSpace(filePath))
	if err != nil {
		fmt.Fprintf((*clientConnection), "No se encuentra un archivo en la ruta")
		return
	}
	defer file.Close()

	for {
		n, err := file.ReadAt(fileBuffer, int64(currentByte))
		currentByte += bufferSize
		fmt.Println(fileBuffer)
		(*clientConnection).Write(fileBuffer[:n])
		if err == io.EOF {
			break
		}
	}
	file.Close()
}

func getDataFromClient(filePath string, clientConnection *net.Conn, bufferSize uint32) {
	var currentByte uint32 = 0

	fileBuffer := make([]byte, bufferSize)

	file, err := os.Create(strings.TrimSpace(filePath))
	if err != nil {
		fmt.Println("Error creating file")
		log.Fatal(err)
	}
	defer file.Close()

	for {
		n, err := (*clientConnection).Read(fileBuffer)
		fileBufferString := string(fileBuffer[:])

		if err == io.EOF || err != nil {
			break
		}
		if fileBufferString == "No se encuentra un archivo en la ruta" {
			fmt.Println("Error al cargar el archivo")
			return
		}
		fmt.Println("File Buffer", fileBuffer)

		_, err = file.WriteAt(fileBuffer[:n], int64(currentByte))

		currentByte += bufferSize

		if err == io.EOF || uint32(n) != bufferSize {
			break
		}
	}
	fmt.Println("Out reading")
	file.Close()
}
