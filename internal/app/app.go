package app

import (
	"fmt"

	"github.com/MSSkowron/GRPCChatter/internal/server"
)

func Run() error {
	if err := server.NewGRPCChatterServer().ListenAndServe(); err != nil {
		return fmt.Errorf("failed to run server: %w", err)
	}

	return nil
}
