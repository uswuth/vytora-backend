// Package httputil provides HTTP utilities for request handling.
package httputil

import (
	"net"
	"net/http"
	"strings"
)

// ClientIP extracts the real client IP from an HTTP request.
// It checks X-Forwarded-For first (for proxied requests), then falls back to RemoteAddr with port stripped.
func ClientIP(r *http.Request) string {
	if fwd := r.Header.Get("X-Forwarded-For"); fwd != "" {
		parts := strings.Split(fwd, ",")
		return strings.TrimSpace(parts[0])
	}
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}