package middleware

import (
	"net/http"

	"github.com/google/uuid"
)

func RequestIDMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		requestId := uuid.New().String()

		r.Header.Set("X-Request-ID", requestId)

		w.Header().Set("X-Request-ID", requestId)

		next.ServeHTTP(w, r)
	})
}
