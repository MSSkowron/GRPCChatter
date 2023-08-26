package app

import (
	"flag"
	"fmt"

	"github.com/MSSkowron/GRPCChatter/internal/server"
	"github.com/MSSkowron/GRPCChatter/internal/services"
)

const (
	secret              = "12345678901234567890123456789012"
	shortCodeLength     = 6
	maxMessageQueueSize = 255
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
	flag.IntVar(&config.MaxMessageQueueSize, "queue_size", maxMessageQueueSize, "GRPCChatter server max message queue size")

	flag.Parse()

	return config
}

func Run() error {
	config := ParseConfig()

	tokenService := services.NewTokenService(secret)
	shortCodeService := services.NewShortCodeService(shortCodeLength)
	clientsRoomsService := services.NewClientsRoomsServiceImpl(config.MaxMessageQueueSize)

	server := server.NewGRPCChatterServer(
		tokenService,
		shortCodeService,
		clientsRoomsService,
		server.WithAddress(config.Address),
		server.WithPort(config.Port),
	)

	if err := server.ListenAndServe(); err != nil {
		return fmt.Errorf("failed to run server: %w", err)
	}

	return nil
}
