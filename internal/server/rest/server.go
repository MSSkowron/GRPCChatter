package rest

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/MSSkowron/GRPCChatter/internal/dto"
	"github.com/MSSkowron/GRPCChatter/internal/service"
	"github.com/MSSkowron/GRPCChatter/pkg/logger"
	"github.com/gorilla/mux"
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
	userService service.UserService
}

// NewServer creates a new Server instance.
func NewServer(userService service.UserService, opts ...ServerOption) *Server {
	server := &Server{
		Server: &http.Server{
			Addr:         DefaultAddress,
			WriteTimeout: DefaultWriteTimeout,
			ReadTimeout:  DefaultReadTimeout,
		},
		userService: userService,
	}

	for _, opt := range opts {
		opt(server)
	}

	server.initRoutes()

	return server
}

// ServerOption is a function signature for providing options to configure the Server.
type ServerOption func(*Server)

// WithAddress is an option to set the server address.
func WithAddress(addr string) ServerOption {
	return func(s *Server) {
		s.Addr = addr
	}
}

// WithReadTimeout is an option to set the read timeout for the server.
func WithReadTimeout(timeout time.Duration) ServerOption {
	return func(s *Server) {
		s.ReadTimeout = timeout
	}
}

// WithWriteTimeout is an option to set the write timeout for the server.
func WithWriteTimeout(timeout time.Duration) ServerOption {
	return func(s *Server) {
		s.WriteTimeout = timeout
	}
}

func (s *Server) initRoutes() {
	r := mux.NewRouter()

	r.Use(s.log)

	r.HandleFunc("/register", s.handleRegister).Methods("POST")
	r.HandleFunc("/login", s.handleLogin).Methods("POST")

	s.Handler = r
}

func (s *Server) log(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ip, endpoint, method := r.RemoteAddr, r.URL.Path, r.Method

		body, err := io.ReadAll(r.Body)
		if err != nil {
			s.respondWithError(w, http.StatusInternalServerError, "Internal server error")
			return
		}
		r.Body = io.NopCloser(bytes.NewBuffer(body))

		logger.Info(fmt.Sprintf("Received request from [%s] to [%s] with method [%s] and body [%s]", ip, endpoint, method, string(body)))

		next.ServeHTTP(w, r)
	})
}

func (s *Server) handleRegister(w http.ResponseWriter, r *http.Request) {
	registerDTO := &dto.UserRegisterDTO{}
	if err := json.NewDecoder(r.Body).Decode(registerDTO); err != nil {
		s.respondWithError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	userDTO, err := s.userService.RegisterUser(r.Context(), registerDTO)
	if err != nil {
		s.respondWithError(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	s.respondWithJSON(w, http.StatusOK, userDTO)
}

func (s *Server) handleLogin(w http.ResponseWriter, r *http.Request) {
	loginDTO := &dto.UserLoginDTO{}
	if err := json.NewDecoder(r.Body).Decode(loginDTO); err != nil {
		s.respondWithError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	tokenDTO, err := s.userService.LoginUser(r.Context(), loginDTO)
	if err != nil {
		s.respondWithError(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	s.respondWithJSON(w, http.StatusOK, tokenDTO)
}

func (s *Server) respondWithError(w http.ResponseWriter, errCode int, errMessage string) {
	s.respondWithJSON(w, errCode, dto.ErrorDTO{Error: errMessage})
}

func (s *Server) respondWithJSON(w http.ResponseWriter, code int, payload any) {
	w.Header().Set("Content-Type", "application/json")

	response, err := json.Marshal(payload)
	if err != nil {
		logger.Error(fmt.Sprintf("Failed to marshall response to JSON: %s ", err))

		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte("Internal server error"))

		return
	}

	w.WriteHeader(code)
	_, _ = w.Write(response)
}
