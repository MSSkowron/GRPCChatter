package app

import (
	"flag"
	"fmt"

	"github.com/MSSkowron/GRPCChatter/internal/config"
	"github.com/MSSkowron/GRPCChatter/internal/server"
	"github.com/MSSkowron/GRPCChatter/internal/services"
)

const (
	defaultConfigFilePath = "./configs/default_config.env"
)

func Run() error {
	configFilePath := flag.String("config", defaultConfigFilePath, "GRPCChatter configuration file path")
	flag.Parse()

	config, err := config.Load(*configFilePath)
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	tokenService := services.NewTokenService(config.Secret)
	shortCodeService := services.NewShortCodeService(config.ShortCodeLength)
	roomService := services.NewRoomService(config.MaxMessageQueueSize)

	server := server.NewGRPCChatterServer(
		tokenService,
		shortCodeService,
		roomService,
		server.WithAddress(config.ServerAddress),
		server.WithPort(config.ServerPort),
	)

	if err := server.ListenAndServe(); err != nil {
		return fmt.Errorf("failed to run server: %w", err)
	}

	return nil
}
