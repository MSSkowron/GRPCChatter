package logger

import (
	"log/slog"
	"os"
)

var logger *slog.Logger

func init() {
	logger = slog.New(slog.NewJSONHandler(os.Stdout, nil))
}

// Info logs the provided message at [InfoLevel].
func Info(msg string) {
	logger.Info(msg)
}

// Debug logs the provided message at [DebugLevel].
func Debug(msg string) {
	logger.Debug(msg)
}

// Error logs the provided message at [ErrorLevel].
func Error(msg string) {
	logger.Error(msg)
}
