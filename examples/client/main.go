package main

import (
	"fmt"
	"log"

	"github.com/MSSkowron/GRPCChatter/pkg/client"
)

func main() {
	c := client.NewClient("alicja", ":5000")
	defer c.Close()

	if err := c.Join(); err != nil {
		log.Fatalln(err)
	}

	if err := c.Send("hello"); err != nil {
		log.Fatalln(err)
	}

	msg, err := c.Receive()
	if err != nil {
		log.Fatalln(err)
	}

	fmt.Printf("[%s]: %s\n", msg.Sender, msg.Body)
}
