package main

import (
	"fmt"

	"github.com/MSSkowron/GRPCChatter/pkg/client"
)

func main() {
	c := client.NewClient("mateusz", ":5000")
	defer c.Close()

	c.Join()

	c.Send("hello")

	fmt.Println(c.Receive())
}
