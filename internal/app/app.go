package app

import (
	"flag"
	"fmt"

	"github.com/MSSkowron/GRPCChatter/internal/server"
)

func Run() error {
	serverAddress := flag.String("address", server.DefaultAddress, "GRPCChatter server listening address")
	serverPort := flag.String("port", server.DefaultPort, "GRPCChatter server listening port")
	serverMaxQueueSize := flag.Int("queue_size", server.DefaultMaxMessageQueueSize, "GRPCChatter server max message queue size")

	if err := server.NewGRPCChatterServer(server.WithAddress(*serverAddress), server.WithPort(*serverPort), server.WithMaxMessageQueueSize(*serverMaxQueueSize)).ListenAndServe(); err != nil {
		return fmt.Errorf("failed to run server: %w", err)
	}

	return nil
}
