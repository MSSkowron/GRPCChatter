package app

import (
	"context"
	"flag"
	"fmt"

	"github.com/MSSkowron/GRPCChatter/internal/config"
	"github.com/MSSkowron/GRPCChatter/internal/database"
	"github.com/MSSkowron/GRPCChatter/internal/repository"
	"github.com/MSSkowron/GRPCChatter/internal/server/grpc"
	"github.com/MSSkowron/GRPCChatter/internal/server/rest"
	"github.com/MSSkowron/GRPCChatter/internal/service"
	"golang.org/x/sync/errgroup"
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

	database, err := database.NewPostgresDatabase(context.Background(), config.DatabaseURL)
	if err != nil {
		return fmt.Errorf("failed to create database: %w", err)
	}
	defer database.Close()

	userRepository := repository.NewUserRepository(database)

	userTokenService := service.NewUserTokenService(config.Secret, config.TokenDuration)
	userService := service.NewUserService(userTokenService, userRepository)
	chatTokenService := service.NewChatTokenService(config.Secret)
	shortCodeService := service.NewShortCodeService(config.ShortCodeLength)
	roomService := service.NewRoomService(config.MaxMessageQueueSize)

	grpcServer := grpc.NewServer(
		chatTokenService,
		shortCodeService,
		roomService,
		grpc.WithAddress(config.GRPCServerAddress),
		grpc.WithPort(config.GRPCServerPort),
	)

	restServer := rest.NewServer(
		userService,
		rest.WithAddress(fmt.Sprintf("%s:%d", config.RESTServerAddress, config.RESTServerPort)),
	)

	g := errgroup.Group{}

	g.Go(func() error {
		if err := restServer.ListenAndServe(); err != nil {
			return fmt.Errorf("failed to run REST server: %w", err)
		}
		return nil
	})

	g.Go(func() error {
		if err := grpcServer.ListenAndServe(); err != nil {
			return fmt.Errorf("failed to run gRPC server: %w", err)
		}
		return nil
	})

	return g.Wait()
}
