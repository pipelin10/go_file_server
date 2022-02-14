package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"strings"
)

func main() {
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

		message, err := bufio.NewReader(c).ReadString('\n')
		if err != nil {
			fmt.Println(err)
			fmt.Println("Close Connection")
			return
		}
		fmt.Print("->:" + message)
		if strings.TrimSpace(string(text)) == "st" {
			fmt.Println("TCP Cliente exting")
			return
		}
	}
}
