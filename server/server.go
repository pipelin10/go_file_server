package main

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"
)

const bufferSize = 1024
const maxQueueLength = 10000

var channels map[string][]net.Conn
var connChannels map[string][]string

type TcpConfig struct {
	Host         string
	Port         int
	MaxIdleConns int
	MaxOpenConn  int
}

type tcpConn struct {
	id   string
	pool *TcpConnPool
	conn net.Conn
}

type connRequest struct {
	connChan chan *tcpConn
	errChan  chan error
}

type TcpConnPool struct {
	host         string
	port         int
	mu           sync.Mutex
	idleConns    map[string]*tcpConn
	numOpen      int
	maxOpenCount int
	maxIdleCount int
	requestChan  chan *connRequest
}

func (p *TcpConnPool) put(c *tcpConn) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.maxIdleCount > 0 && p.maxIdleCount > len(p.idleConns) {
		p.idleConns[c.id] = c
	} else {
		c.conn.Close()
		c.pool.numOpen--
	}
}

func (p *TcpConnPool) get() (*tcpConn, error) {
	p.mu.Lock()

	numIdle := len(p.idleConns)
	if numIdle > 0 {
		for _, c := range p.idleConns {
			delete(p.idleConns, c.id)
			p.mu.Unlock()
			return c, nil
		}
	}

	if p.maxOpenCount > 0 && p.numOpen >= p.maxOpenCount {
		req := &connRequest{
			connChan: make(chan *tcpConn, 1),
			errChan:  make(chan error, 1),
		}

		p.requestChan <- req

		p.mu.Unlock()

		select {
		case tcpConn := <-req.connChan:
			return tcpConn, nil
		case err := <-req.errChan:
			return nil, err
		}
	}

	p.numOpen++
	p.mu.Lock()

	newTcPConn, err := p.openNewTcpConnection()
	if err != nil {
		p.mu.Lock()
		p.numOpen--
		p.mu.Unlock()
		return nil, err
	}

	return newTcPConn, nil
}

func (p *TcpConnPool) openNewTcpConnection() (*tcpConn, error) {
	addr := fmt.Sprintf("%s:%d", p.host, p.port)

	c, err := net.Dial("tcp4", addr)
	if err != nil {
		return nil, err
	}

	return &tcpConn{
		id:   fmt.Sprintf("%v", time.Now().UnixNano()),
		conn: c,
		pool: p,
	}, nil
}

func (p *TcpConnPool) handleConnectionRequest() {
	for req := range p.requestChan {
		var (
			requestDone = false
			hasTimeOut  = false

			timeoutChan = time.After(3 * time.Second)
		)

		for {
			if requestDone || hasTimeOut {
				break
			}

			select {
			case <-timeoutChan:
				hasTimeOut = true
				req.errChan <- errors.New("connection request timeout")
			default:
				p.mu.Lock()

				numIdle := len(p.idleConns)
				if numIdle > 0 {
					for _, c := range p.idleConns {
						delete(p.idleConns, c.id)
						p.mu.Unlock()
						req.connChan <- c
						requestDone = true
						break
					}
				} else if p.maxOpenCount > 0 && p.numOpen < p.maxOpenCount {
					p.numOpen++
					p.mu.Unlock()

					c, err := p.openNewTcpConnection()
					if err != nil {
						p.mu.Lock()
						p.numOpen--
						p.mu.Unlock()
					} else {
						req.connChan <- c
						requestDone = true
					}
				} else {
					p.mu.Unlock()
				}
			}
		}
	}

}

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

func CreateTcpConnPool(cfg *TcpConfig) (*TcpConnPool, error) {
	pool := &TcpConnPool{
		host:         cfg.Host,
		port:         cfg.Port,
		idleConns:    make(map[string]*tcpConn),
		requestChan:  make(chan *connRequest, maxQueueLength),
		maxOpenCount: cfg.MaxOpenConn,
		maxIdleCount: cfg.MaxIdleConns,
	}

	go pool.handleConnectionRequest()

	return pool, nil
}

func handleConnection(c net.Conn) {
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
		if command == "STOP" {
			fmt.Printf("Closing connection with %s\n", c.RemoteAddr().String())
			eraseConnChannels(c)
			break
		} else if command == "get" {
			filePath := ".\\files_to_send_server\\" + splitMessage[1]
			send_data_to_client(filePath, c)
		} else if command == "send" {
			fileName := splitMessage[1]
			filePath := ".\\files_recieved_server\\" + fileName
			channel := splitMessage[2]
			get_data_from_client(filePath, c)
			fmt.Print(channel)
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
			connChannels[c.RemoteAddr().String()] = append(connChannels[c.RemoteAddr().String()], channel)
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
	connChannels = make(map[string][]string)

	cfg := &TcpConfig{
		Host:         "localhost",
		Port:         8080,
		MaxOpenConn:  1000,
		MaxIdleConns: 1000,
	}

	pool, err := CreateTcpConnPool(cfg)
	if err != nil {
		log.Fatal(err)
	}
	go pool.handleConnectionRequest()
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
		defer c.Close()
		go handleConnection(c)
	}
}
