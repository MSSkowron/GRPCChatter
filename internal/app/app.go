package app

import (
	"flag"
	"fmt"

	"github.com/MSSkowron/GRPCChatter/internal/server"
)

const (
	secret = "12345678901234567890123456789012"
)

// Config holds configuration options for the GRPCChatter server.
type Config struct {
	Address             string
	Port                int
	MaxMessageQueueSize int
}

// ParseConfig parses command line arguments and returns a Config struct.
func ParseConfig() Config {
	var config Config

	flag.StringVar(&config.Address, "address", server.DefaultAddress, "GRPCChatter server listening address")
	flag.IntVar(&config.Port, "port", server.DefaultPort, "GRPCChatter server listening port")
	flag.IntVar(&config.MaxMessageQueueSize, "queue_size", server.DefaultMaxMessageQueueSize, "GRPCChatter server max message queue size")

	flag.Parse()

	return config
}

func Run() error {
	config := ParseConfig()

	server := server.NewGRPCChatterServer(
		secret,
		server.WithAddress(config.Address),
		server.WithPort(config.Port),
		server.WithMaxMessageQueueSize(config.MaxMessageQueueSize),
	)

	if err := server.ListenAndServe(); err != nil {
		return fmt.Errorf("failed to run server: %w", err)
	}

	return nil
}
