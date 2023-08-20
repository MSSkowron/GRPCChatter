package main

import "github.com/MSSkowron/GRPCChatter/internal/app"

func main() {
	if err := app.Run(); err != nil {
		panic(err)
	}
}
