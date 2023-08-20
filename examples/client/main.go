package main

import (
	"fmt"
	"log"

	"github.com/MSSkowron/GRPCChatter/pkg/client"
)

func main() {
	c, err := client.NewClient("mateusz", ":5000")
	if err != nil {
		log.Fatalln(err)
	}
	defer c.Close()

	c.Join()

	c.Send("hello")

	fmt.Println(c.Receive())
}
