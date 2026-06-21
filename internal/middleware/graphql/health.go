package graphql

import (
	"context"
	"net/http"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

var startTime = time.Now()

func HealthCheckHandler(allowedIPs []string) http.HandlerFunc {
	allowed := map[string]bool{}
	for _, ip := range allowedIPs {
		allowed[ip] = true
	}
	allowed["127.0.0.1"] = true
	allowed["::1"] = true

	return func(w http.ResponseWriter, r *http.Request) {
		ip := r.RemoteAddr
		if fwd := r.Header.Get("X-Forwarded-For"); fwd != "" {
			ip = fwd
		}
		if !allowed[ip] {
			http.Error(w, "forbidden", http.StatusForbidden)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"status":"ok","uptime":"` + time.Since(startTime).String() + `"}`))
	}
}

func ReadinessHandler(allowedIPs []string, pool *pgxpool.Pool) http.HandlerFunc {
	allowed := map[string]bool{}
	for _, ip := range allowedIPs {
		allowed[ip] = true
	}
	allowed["127.0.0.1"] = true
	allowed["::1"] = true

	return func(w http.ResponseWriter, r *http.Request) {
		ip := r.RemoteAddr
		if fwd := r.Header.Get("X-Forwarded-For"); fwd != "" {
			ip = fwd
		}
		if !allowed[ip] {
			http.Error(w, "forbidden", http.StatusForbidden)
			return
		}
		ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
		defer cancel()
		dbStatus := "ok"
		if err := pool.Ping(ctx); err != nil {
			dbStatus = "down"
		}
		status := "ready"
		code := http.StatusOK
		if dbStatus != "ok" {
			status = "not ready"
			code = http.StatusServiceUnavailable
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(code)
		w.Write([]byte(`{"status":"` + status + `","checks":{"database":"` + dbStatus + `"}}`))
	}
}