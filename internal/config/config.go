package config

import (
	"fmt"
	"time"

	"github.com/spf13/viper"
)

// Config stores configuration values for the application.
// These values can be read from a configuration file or environment variables.
type Config struct {
	// DatabaseURL is a connection string to the database.
	DatabaseURL string `mapstructure:"DATABASE_URL"`
	// GRPCServerAddress is the IP address where the gRPC server will listen.
	GRPCServerAddress string `mapstructure:"GRPC_SERVER_ADDRESS"`
	// GRPCServerPort is the port on which the gRPC server will listen.
	GRPCServerPort int `mapstructure:"GRPC_SERVER_PORT"`
	// RESTServerAddress is the IP address where the REST server will listen.
	RESTServerAddress string `mapstructure:"REST_SERVER_ADDRESS"`
	// RESTServerPort is the port on which the REST server will listen.
	RESTServerPort int `mapstructure:"REST_SERVER_PORT"`
	// Secret is a secret key used for JWT token signing and validation.
	Secret string `mapstructure:"SECRET"`
	// ShortCodeLength is the length of generated room short codes.
	ShortCodeLength int `mapstructure:"SHORT_CODE_LENGTH"`
	// MaxMessageQueueSize is the maximum size of the message queue.
	MaxMessageQueueSize int `mapstructure:"MAX_MESSAGE_QUEUE_SIZE"`
	// TokenDuration is a duration for which the JWT token is valid.
	TokenDuration time.Duration `mapstructure:"TOKEN_DURATION"`
}

// Load loads configuration settings from a specified file or environment variables.
// If both a configuration file and environment variables are used, environment variables take precedence.
func Load(filePath string) (*Config, error) {
	viper.SetConfigFile(filePath)
	viper.SetConfigType("env")
	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	config := &Config{}
	if err := viper.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	return config, nil
}
