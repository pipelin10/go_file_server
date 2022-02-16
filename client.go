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

func main() {
	const bufferSize = 1024
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
			currentByte := 0
			fileName := ".\\files_recieved\\" + text_split[1]

			fileBuffer := make([]byte, bufferSize)

			file, err := os.Create(strings.TrimSpace(fileName))
			if err != nil {
				fmt.Println("Error creating file")
				log.Fatal(err)
			}
			defer file.Close()
			// defer c.Close()

			// c.Write([]byte("get " + fileName + "\n"))
			// fmt.Println("Before reading")
			// c.Read(fileBuffer)
			// fmt.Println("File Buffer", fileBuffer)

			for {
				fmt.Println("Inside reading")
				fmt.Println("Before reading")
				n, err := c.Read(fileBuffer)
				fileBufferString := string(fileBuffer[:])
				fmt.Println("File Buffer String", fileBufferString)
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
