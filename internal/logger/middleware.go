package logger

import (
	"net/http"
	"time"

	"github.com/uswuth/vytora-backend/internal/startup"
)

// RequestLogger logs HTTP requests in a minimal table format
func RequestLogger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
	rec := startup.NewResponseRecorder(w)
		next.ServeHTTP(rec, r)
		startup.LogRequest(rec, r, start)
	})
}