package main

import (
	"bufio"
	"fmt"
	"io"
	"math/rand"
	"net"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

const bufferSize = 1024

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

func handleConnection(c net.Conn) {
	fmt.Printf("Serving %s\n", c.RemoteAddr().String())
	for {
		netData, err := bufio.NewReader(c).ReadString('\n')
		if err != nil {
			fmt.Println(err)
			return
		}
		splitMessage := strings.Split(string(netData), " ")
		temp := strings.TrimSpace(splitMessage[0])
		fmt.Println(splitMessage)
		if temp == "STOP" {
			fmt.Printf("Closing connection with %s\n", c.RemoteAddr().String())
			break
		} else if temp == "get" {
			fileName := ".\\files_to_send\\" + splitMessage[1]
			currentByte := 0

			fileBuffer := make([]byte, bufferSize)

			file, err := os.Open(strings.TrimSpace(fileName))
			fmt.Println("File:", file)
			if err != nil {
				fmt.Fprintf(c, "No se encuentra un archivo en la ruta")
				continue
			}
			defer file.Close()

			for {
				n, err := file.ReadAt(fileBuffer, int64(currentByte))
				currentByte += bufferSize
				fmt.Println(fileBuffer)
				c.Write(fileBuffer[:n])
				fmt.Println("Sent", n, "bytes")
				if err == io.EOF {
					break
				}
			}
			fmt.Println("Closing File")
			file.Close()
			netData, err = bufio.NewReader(c).ReadString('\n')
			fmt.Println("Response from server:", netData)
			if err != nil {
				fmt.Println(err)
				return
			}
			netData, err = bufio.NewReader(c).ReadString('\n')
			fmt.Println("Response from server:", netData)
			if err != nil {
				fmt.Println(err)
				return
			}
		} else {
			result := strconv.Itoa(rand.Intn(100-1)+1) + "\n"
			c.Write([]byte(string(result)))
		}

	}
	c.Close()
}

func main() {
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
