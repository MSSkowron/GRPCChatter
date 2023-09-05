package rest

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/MSSkowron/GRPCChatter/pkg/logger"
	"github.com/google/uuid"
)

func (s *Server) logMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := uuid.New().String()

		clientIP := getClientIP(r)
		endpoint := r.URL.Path
		method := r.Method

		body, err := getRequestBody(r)
		if err != nil {
			s.respondWithError(w, http.StatusInternalServerError, ErrMsgInternalServerError)
			return
		}

		logger.Info(fmt.Sprintf("Received request [ID: %s] from [%s] to [%s] with method [%s] and body [%s]", id, clientIP, endpoint, method, string(body)))

		r = r.WithContext(context.WithValue(r.Context(), contextKeyReqID, id))

		next.ServeHTTP(w, r)
	})
}

func getClientIP(r *http.Request) string {
	ip := r.Header.Get("X-Forwarded-For")
	if ip == "" {
		ip = r.RemoteAddr
	}

	idx := strings.Index(ip, ":")
	if idx != -1 {
		ip = ip[:idx]
	}
	return ip
}

func getRequestBody(r *http.Request) ([]byte, error) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		return nil, err
	}
	r.Body = io.NopCloser(bytes.NewBuffer(body))
	return body, nil
}
