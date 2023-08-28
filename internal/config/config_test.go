package config

import (
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestLoadValid(t *testing.T) {
	configFile := createTempConfigFile(t)
	defer os.Remove(configFile)

	cfg, err := Load(configFile)
	require.NoError(t, err)

	require.Equal(t, "127.0.0.1", cfg.ServerAddress)
	require.Equal(t, 5000, cfg.ServerPort)
	require.Equal(t, "123ABC", cfg.Secret)
	require.Equal(t, 6, cfg.ShortCodeLength)
	require.Equal(t, 255, cfg.MaxMessageQueueSize)
}

func TestLoadConfigInvalidPath(t *testing.T) {
	configFile := createTempConfigFile(t)
	defer os.Remove(configFile)

	_, err := Load("invalid_path_config.env")
	require.ErrorIs(t, err, os.ErrNotExist)
}

func createTempConfigFile(t *testing.T) string {
	configFile := "temp_config.env"
	file, err := os.Create(configFile)
	require.NoError(t, err)
	defer file.Close()

	_, err = file.WriteString("SERVER_ADDRESS=127.0.0.1\n")
	require.NoError(t, err)

	_, err = file.WriteString("SERVER_PORT=5000\n")
	require.NoError(t, err)

	_, err = file.WriteString("SECRET=123ABC\n")
	require.NoError(t, err)

	_, err = file.WriteString("SHORT_CODE_LENGTH=6\n")
	require.NoError(t, err)

	_, err = file.WriteString("MAX_MESSAGE_QUEUE_SIZE=255\n")
	require.NoError(t, err)

	return configFile
}
