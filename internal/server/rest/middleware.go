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
		requestID := uuid.New().String()

		clientIP := getClientIP(r)
		endpoint := r.URL.Path
		httpMethod := r.Method

		requestBody, err := getRequestBody(r)
		if err != nil {
			s.respondWithError(w, http.StatusInternalServerError, ErrMsgInternalServerError)
			return
		}

		logMessage := fmt.Sprintf(
			"Received request [ID: %s] from [ClientIP: %s] to [Endpoint: %s] with [HTTP Method: %s] and [Request Body: %s]",
			requestID, clientIP, endpoint, httpMethod, string(requestBody),
		)
		logger.Info(logMessage)

		r = r.WithContext(context.WithValue(r.Context(), contextKeyReqID, requestID))

		next.ServeHTTP(w, r)
	})
}

func getClientIP(r *http.Request) string {
	ip := r.Header.Get("X-Forwarded-For")
	if ip == "" {
		ip = r.RemoteAddr
	}

	colonIndex := strings.Index(ip, ":")
	if colonIndex != -1 {
		ip = ip[:colonIndex]
	}
	return ip
}

func getRequestBody(r *http.Request) ([]byte, error) {
	requestBody, err := io.ReadAll(r.Body)
	if err != nil {
		return nil, err
	}
	r.Body = io.NopCloser(bytes.NewBuffer(requestBody))
	return requestBody, nil
}
