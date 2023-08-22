package main

import (
	"os"

	"github.com/MSSkowron/GRPCChatter/internal/app"
	"github.com/MSSkowron/GRPCChatter/pkg/logger"
)

func main() {
	if err := app.Run(); err != nil {
		logger.Error(err.Error())
		os.Exit(1)
	}
}
