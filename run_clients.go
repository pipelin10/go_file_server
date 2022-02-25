package main

import (
	"fmt"
	"time"
)

func get_client() {
	run_client()
}

func main() {
	for i := 0; i < 10000; i++ {
		fmt.Println(i)
		go get_client()
		time.Sleep(1 * time.Millisecond)
	}
}
