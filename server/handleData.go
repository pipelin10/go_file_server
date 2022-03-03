package main

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"strings"
)

func sendDataToClient(filePath string, conn net.Conn, bufferSize uint32) {
	var currentByte uint32 = 0

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

func getDataFromClient(filePath string, conn net.Conn, bufferSize uint32) {
	var currentByte uint32 = 0

	fileBuffer := make([]byte, bufferSize)

	file, err := os.Create(strings.TrimSpace(filePath))
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

		if err == io.EOF || uint32(n) != bufferSize {
			fmt.Println("Algarete!!!")
			break
		}
	}
	// c.Write([]byte("get " + filePath + "\n"))
	fmt.Println("Out reading")
	file.Close()
}
