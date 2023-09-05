package rest

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/MSSkowron/GRPCChatter/internal/dto"
	"github.com/MSSkowron/GRPCChatter/internal/service"
	"github.com/MSSkowron/GRPCChatter/pkg/logger"
	"github.com/MSSkowron/GRPCChatter/pkg/validation"
	"github.com/gorilla/mux"
)

type contextKey string

const (
	// DefaultPort is the default port the server listens on.
	DefaultPort = 5000
	// DefaultAddress is the default address the server listens on.
	DefaultAddress = ""
	// DefaultWriteTimeout is the default write timeout for server responses.
	DefaultWriteTimeout = 15 * time.Second
	// DefaultReadTimeout is the default read timeout for incoming requests.
	DefaultReadTimeout = 15 * time.Second

	contextKeyReqID = contextKey("reqID")

	// ErrMsgUnauthorized is a http response body message for unauthorized status code.
	ErrMsgUnauthorized = "Unauthorized"
	// ErrMsgBadRequestInvalidRequestBody is a http response body message for bad request status code.
	ErrMsgBadRequestInvalidRequestBody = "Invalid request body"
	// ErrMsgInternalServerError is a http response body message for internal server error status code.
	ErrMsgInternalServerError = "Internal server error"
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

	r.Use(s.logMiddleware)

	r.HandleFunc("/register", s.handleRegister).Methods("POST")
	r.HandleFunc("/login", s.handleLogin).Methods("POST")

	s.Handler = r
}

func (s *Server) handleRegister(w http.ResponseWriter, r *http.Request) {
	registerDTO := &dto.UserRegisterDTO{}
	if err := json.NewDecoder(r.Body).Decode(registerDTO); err != nil {
		s.respondWithError(w, http.StatusBadRequest, ErrMsgBadRequestInvalidRequestBody)
		return
	}

	userDTO, err := s.userService.RegisterUser(r.Context(), registerDTO)
	if err != nil {
		switch {
		case errors.Is(err, validation.ErrInvalidUsername):
			s.respondWithError(w, http.StatusBadRequest, fmt.Sprintf("%s:%s", ErrMsgBadRequestInvalidRequestBody, err))
		case errors.Is(err, validation.ErrInvalidPassword):
			s.respondWithError(w, http.StatusBadRequest, fmt.Sprintf("%s:%s", ErrMsgBadRequestInvalidRequestBody, err))
		case errors.Is(err, service.ErrUserAlreadyExists):
			s.respondWithError(w, http.StatusBadRequest, fmt.Sprintf("%s:%s", ErrMsgBadRequestInvalidRequestBody, err))
		default:
			s.respondWithError(w, http.StatusInternalServerError, ErrMsgInternalServerError)
		}
		return
	}

	s.respondWithJSON(w, http.StatusOK, userDTO)
}

func (s *Server) handleLogin(w http.ResponseWriter, r *http.Request) {
	loginDTO := &dto.UserLoginDTO{}
	if err := json.NewDecoder(r.Body).Decode(loginDTO); err != nil {
		s.respondWithError(w, http.StatusBadRequest, ErrMsgBadRequestInvalidRequestBody)
		return
	}

	tokenDTO, err := s.userService.LoginUser(r.Context(), loginDTO)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrInvalidCredentials):
			s.respondWithError(w, http.StatusUnauthorized, fmt.Sprintf("%s:%s", ErrMsgUnauthorized, err))
		default:
			s.respondWithError(w, http.StatusInternalServerError, ErrMsgInternalServerError)
		}
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
		if _, err := w.Write([]byte(ErrMsgInternalServerError)); err != nil {
			logger.Error(fmt.Sprintf("Failed to respond: %s", err))
		}

		return
	}

	w.WriteHeader(code)
	if _, err := w.Write(response); err != nil {
		logger.Error(fmt.Sprintf("Failed to respond: %s", err))
	}
}
