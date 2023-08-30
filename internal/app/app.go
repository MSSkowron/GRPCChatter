package app

import (
	"flag"
	"fmt"

	"github.com/MSSkowron/GRPCChatter/internal/config"
	"github.com/MSSkowron/GRPCChatter/internal/server/grpc"
	"github.com/MSSkowron/GRPCChatter/internal/service"
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

	chatTokenService := service.NewChatTokenService(config.Secret)
	shortCodeService := service.NewShortCodeService(config.ShortCodeLength)
	roomService := service.NewRoomService(config.MaxMessageQueueSize)

	server := grpc.NewServer(
		chatTokenService,
		shortCodeService,
		roomService,
		grpc.WithAddress(config.ServerAddress),
		grpc.WithPort(config.ServerPort),
	)

	if err := server.ListenAndServe(); err != nil {
		return fmt.Errorf("failed to run server: %w", err)
	}

	return nil
}
