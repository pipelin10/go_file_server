package main

import (
	"bufio"
	"bytes"
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"strings"
	"sync"
)

var channels map[string][]net.Conn
var connChannels map[string][]string

type TcpConfig struct {
	Host        string
	Port        int
	MaxOpenConn int
}

type tcpConn struct {
	Id   string
	Pool *TcpConnPool
	Conn net.Conn
}

type TcpConnPool struct {
	Host        string
	Port        int
	Mu          sync.Mutex
	Connections []*tcpConn
	NumOpen     int
	MaxOpenConn int
}

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

func eraseConnChannels(c net.Conn) {
	remoteAddress := c.RemoteAddr().String()
	fmt.Println(connChannels[remoteAddress])
	for _, channel := range connChannels[remoteAddress] {
		for index, conn := range channels[channel] {
			connAddress := conn.RemoteAddr().String()
			if connAddress == remoteAddress {
				channels[channel] = append(channels[channel][:index], channels[channel][index+1:]...)
				break
			}
		}
	}
}

func handleConnection(c net.Conn) {
	const STOP string = "STOP"
	const GET string = "get"
	const SEND string = "send"
	const SUBSCRIBE string = "subscribe"
	const bufferSize uint32 = 1024

	fmt.Printf("Serving %s\n", c.RemoteAddr().String())
	defer c.Close()
	for {
		netData, err := bufio.NewReader(c).ReadString('\n')
		if err != nil {
			fmt.Println(err)
			eraseConnChannels(c)
			return
		}
		splitMessage := strings.Split(string(netData), " ")
		command := strings.TrimSpace(splitMessage[0])
		fmt.Println(splitMessage)
		if command == STOP {
			fmt.Printf("Closing connection with %s\n", c.RemoteAddr().String())
			eraseConnChannels(c)
			break
		} else if command == GET {
			if len(splitMessage) == 1 {
				fmt.Fprintf(c, "Please specify a filename\n")
				continue
			}
			fileName := splitMessage[1]
			filePath := ".\\files_to_send_server\\" + fileName
			sendDataToClient(filePath, c, bufferSize)
		} else if command == SEND {
			if len(splitMessage) == 1 {
				fmt.Fprintf(c, "Please specify a filename\n")
				continue
			} else if len(splitMessage) == 2 {
				fmt.Fprintf(c, "Please specify a channel\n")
				continue
			}
			fileName := splitMessage[1]
			channel := splitMessage[2]
			filePath := ".\\files_recieved_server\\" + fileName
			getDataFromClient(filePath, c, bufferSize)
			fmt.Print(channel)
			ipHostClientSending := c.RemoteAddr().String()
			for _, conn := range channels[channel] {
				ipHostCienteReceiving := conn.RemoteAddr().String()
				if ipHostClientSending != ipHostCienteReceiving {
					fmt.Fprintf(conn, "send %s\n", fileName)
					sendDataToClient(filePath, conn, bufferSize)
				}
			}
		} else if command == SUBSCRIBE {
			if len(splitMessage) == 1 {
				fmt.Fprintf(c, "Please specify a channel\n")
				continue
			}
			channel := splitMessage[1]
			channels[channel] = append(channels[channel], c)
			connChannels[c.RemoteAddr().String()] = append(connChannels[c.RemoteAddr().String()], channel)
			for chanMap, connArray := range channels {
				fmt.Print(chanMap)
				for conn := range connArray {
					fmt.Print(" ", connArray[conn].RemoteAddr().String())
				}
				fmt.Printf("\n")
			}
		} else {
			fmt.Fprintf(c, "Please specify a valid command\n")
		}
	}
	c.Close()
}

func InitConfig() (*TcpConfig, error) {
	PortConfig := flag.Int("port", 8080, "Listen port")
	HostConfig := flag.String("host", "localhost", "Listen host")
	MaxOpenConnConfig := flag.Int("open", 1000, "Max number of tcp connections")

	flag.Parse()

	config := TcpConfig{
		Host:        *HostConfig,
		Port:        *PortConfig,
		MaxOpenConn: *MaxOpenConnConfig,
	}

	return &config, nil
}

func CreateTcpPoolConn(config *TcpConfig) (*TcpConnPool, error) {
	pool := &TcpConnPool{
		NumOpen:     0,
		Port:        config.Port,
		Host:        config.Host,
		MaxOpenConn: config.MaxOpenConn,
		Connections: make([]*tcpConn, 0),
	}

	return pool, nil
}

func main() {
	channels = make(map[string][]net.Conn)
	connChannels = make(map[string][]string)

	config, err := InitConfig()

	if err != nil {
		log.Fatal(err)
	}

	tcpPool, err := CreateTcpPoolConn(config)

	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(tcpPool)

	address := fmt.Sprintf("%s:%d", tcpPool.Host, tcpPool.Port)

	listener, err := net.Listen("tcp4", address)

	if err != nil {
		fmt.Println(err)
		return
	}
	defer listener.Close()

	for {
		connectionClient, err := listener.Accept()

		if err != nil {
			fmt.Println(err)
			return
		}
		defer connectionClient.Close()

		if tcpPool.NumOpen >= tcpPool.MaxOpenConn {
			fmt.Fprintf(connectionClient, "Can't establish a connection with server")
			continue
		}

		tcpPool.NumOpen++

		go handleConnection(connectionClient)
	}
}
