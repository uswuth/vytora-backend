package graphql

import (
	"context"
	"net/http"
	"strings"

	"github.com/uswuth/vytora-backend/internal/services"
)

type contextKey string

const (
	UserContextKey    contextKey = "user"
	RawTokenContextKey contextKey = "rawToken"
)

func AuthMiddleware(jwtService *services.JWTService) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				next.ServeHTTP(w, r)
				return
			}
			parts := strings.SplitN(authHeader, " ", 2)
			if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
				http.Error(w, "invalid authorization format", http.StatusUnauthorized)
				return
			}
			claims, err := jwtService.ValidateToken(parts[1])
			if err != nil {
				http.Error(w, "invalid or expired token", http.StatusUnauthorized)
				return
			}
			ctx := context.WithValue(r.Context(), UserContextKey, claims)
			ctx = context.WithValue(ctx, RawTokenContextKey, parts[1])
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func GetClaims(ctx context.Context) (*services.Claims, bool) {
	c, ok := ctx.Value(UserContextKey).(*services.Claims)
	return c, ok
}

func GetRawToken(ctx context.Context) (string, bool) {
	t, ok := ctx.Value(RawTokenContextKey).(string)
	return t, ok
}