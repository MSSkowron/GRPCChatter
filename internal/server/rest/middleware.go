package rest

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"

	"github.com/MSSkowron/GRPCChatter/pkg/logger"
	"github.com/google/uuid"
)

func (s *Server) logMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id, ip, endpoint, method := uuid.New().String(), r.RemoteAddr, r.URL.Path, r.Method

		body, err := io.ReadAll(r.Body)
		if err != nil {
			s.respondWithError(w, http.StatusInternalServerError, ErrMsgInternalServerError)
			return
		}
		r.Body = io.NopCloser(bytes.NewBuffer(body))

		logger.Info(fmt.Sprintf("Received request [ID: %s] from [%s] to [%s] with method [%s] and body [%s]", id, ip, endpoint, method, string(body)))

		r = r.WithContext(context.WithValue(r.Context(), contextKeyReqID, id))

		next.ServeHTTP(w, r)
	})
}
