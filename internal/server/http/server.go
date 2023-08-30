package http

import (
	"net/http"
	"time"

	"github.com/MSSkowron/GRPCChatter/internal/service"
)

const (
	// DefaultPort is the default port the server listens on.
	DefaultPort = 5000
	// DefaultAddress is the default address the server listens on.
	DefaultAddress = ""
	// DefaultWriteTimeout is the default write timeout for server responses.
	DefaultWriteTimeout = 15 * time.Second
	// DefaultReadTimeout is the default read timeout for incoming requests.
	DefaultReadTimeout = 15 * time.Second
)

// Server represents a gRPC server.
type Server struct {
	*http.Server
	userService  service.UserService
	tokenService service.UserTokenService
}
