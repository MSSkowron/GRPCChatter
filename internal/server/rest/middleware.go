package rest

import (
	"bytes"
	"fmt"
	"io"
	"net/http"

	"github.com/MSSkowron/GRPCChatter/pkg/logger"
)

func (s *Server) logMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ip, endpoint, method := r.RemoteAddr, r.URL.Path, r.Method

		body, err := io.ReadAll(r.Body)
		if err != nil {
			s.respondWithError(w, http.StatusInternalServerError, ErrMsgInternalServerError)
			return
		}
		r.Body = io.NopCloser(bytes.NewBuffer(body))

		logger.Info(fmt.Sprintf("Received request from [%s] to [%s] with method [%s] and body [%s]", ip, endpoint, method, string(body)))

		next.ServeHTTP(w, r)
	})
}
