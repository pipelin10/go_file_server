package main

import (
	"flag"
	"log"
	"net"
	"sync"
)

type TcpConfig struct {
	Host        string
	Port        int
	MaxOpenConn int
}

type TcpConn struct {
	Id   string
	Pool *TcpConnPool
	Conn net.Conn
}

type TcpConnPool struct {
	Host        string
	Port        int
	Mu          sync.Mutex
	Connections []*TcpConn
	NumOpen     int
	MaxOpenConn int
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
		Connections: make([]*TcpConn, 0),
	}

	return pool, nil
}

func InitPool() (*TcpConnPool, error) {
	config, err := InitConfig()

	if err != nil {
		log.Fatal(err)
	}

	tcpPool, err := CreateTcpPoolConn(config)

	if err != nil {
		log.Fatal(err)
	}

	return tcpPool, nil
}
