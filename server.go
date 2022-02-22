package main

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net"
	"net/http"
	"os"
	"strings"
	"time"
)

const bufferSize = 1024

var channels map[string][]net.Conn

func helloHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/hello" {
		http.Error(w, "404 not found.", http.StatusNotFound)
		return
	}

	if r.Method != "GET" {
		http.Error(w, "Method is not supported.", http.StatusNotFound)
		return
	}

	fmt.Fprintf(w, "Hello World!")
}

func formHandler(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		fmt.Fprintf(w, "ParseForm() err: %v", err)
		return
	}
	fmt.Fprintf(w, "POST request successful")
	name := r.FormValue("name")
	address := r.FormValue("address")

	fmt.Fprintf(w, "Name = %s\n", name)
	fmt.Fprintf(w, "Address = %s\n", address)
}

func send_data_to_client(filePath string, conn net.Conn) {
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

func get_data_from_client(filePath string, conn net.Conn) {
	const bufferSize = 1024
	currentByte := 0

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

		if err == io.EOF || n != bufferSize {
			fmt.Println("Algarete!!!")
			break
		}
	}
	// c.Write([]byte("get " + filePath + "\n"))
	fmt.Println("Out reading")
	file.Close()
}

func handleConnection(c net.Conn) {
	fmt.Printf("Serving %s\n", c.RemoteAddr().String())
	defer c.Close()
	for {
		netData, err := bufio.NewReader(c).ReadString('\n')
		if err != nil {
			fmt.Println(err)
			return
		}
		splitMessage := strings.Split(string(netData), " ")
		command := strings.TrimSpace(splitMessage[0])
		fmt.Println(splitMessage)
		if command == "STOP" {
			fmt.Printf("Closing connection with %s\n", c.RemoteAddr().String())
			break
		} else if command == "get" {
			filePath := ".\\files_to_send_server\\" + splitMessage[1]
			send_data_to_client(filePath, c)
		} else if command == "send" {
			fileName := splitMessage[1]
			filePath := ".\\files_recieved_server\\" + fileName
			channel := splitMessage[2]
			get_data_from_client(filePath, c)
			fmt.Printf(channel)
			ipHostClientSending := c.RemoteAddr().String()
			for _, conn := range channels[channel] {
				ipHostCienteReceiving := conn.RemoteAddr().String()
				if ipHostClientSending != ipHostCienteReceiving {
					fmt.Fprintf(conn, "send %s\n", fileName)
					send_data_to_client(filePath, conn)
				}
			}
		} else if command == "subscribe" {
			channel := splitMessage[1]
			channels[channel] = append(channels[channel], c)
			for chanMap, connArray := range channels {
				fmt.Print(chanMap)
				for conn := range connArray {
					fmt.Print(" ", connArray[conn].RemoteAddr().String())
				}
				fmt.Printf("\n")
			}
		} else {
			fmt.Fprintf(c, "Please specify a command\n")
		}
	}
	c.Close()
}

func main() {
	channels = make(map[string][]net.Conn)
	// fileServer := http.FileServer(http.Dir("./static"))
	// http.Handle("/", fileServer)
	// http.HandleFunc("/hello", helloHandler)
	// http.HandleFunc("/form", formHandler)

	// fmt.Printf("Starting Sever at Port 8080")
	// if err := http.ListenAndServe(":8080", nil); err != nil {
	// 	log.Fatal(err)
	// }

	arguments := os.Args
	if len(arguments) == 1 {
		fmt.Println("Please provide a PORT number!")
		return
	}

	PORT := ":" + arguments[1]
	l, err := net.Listen("tcp4", PORT)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer l.Close()
	rand.Seed(time.Now().Unix())

	for {
		c, err := l.Accept()
		if err != nil {
			fmt.Println(err)
			return
		}
		go handleConnection(c)
	}
}
