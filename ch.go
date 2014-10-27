// Channel playground to try out different patterns
package main

import (
	"fmt"
	"time"
)

func pinger(c chan<- string, str string, delay int) {
	duration := time.Duration(delay) * time.Millisecond
	for {
		c <- str
		time.Sleep(duration)
	}
}

func printer(c <-chan string) {
	for {
		msg := <-c
		fmt.Println(msg)
	}
}

func main() {
	c := make(chan string)

	go pinger(c, "ping", 100)
	go pinger(c, "pong", 250)
	go printer(c)

	var input string
	fmt.Scanln(&input)
}
