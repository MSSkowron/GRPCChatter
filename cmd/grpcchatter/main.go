package main

import (
	"log"

	"github.com/MSSkowron/GRPCChatter/internal/app"
)

func main() {
	if err := app.Run(); err != nil {
		log.Fatalln(err)
	}
}
