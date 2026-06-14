package middleware

import (
	"net/http"

	"github.com/uswuth/vytora-backend/internal/services"
)

func RequirePermission(permission string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			claims, ok := r.Context().Value(UserContextKey).(*services.Claims)
			if !ok {
				http.Error(w, `{"error":"unauthorized"}`, http.StatusUnauthorized)
				return
			}
			allowedPermission, exists := RolePermissions[claims.Role]
			if !exists {
				http.Error(w, `"error":"forbidden`, http.StatusForbidden)
				return
			}
			for _, p := range allowedPermission {
				if p == permission {
					next.ServeHTTP(w, r)
					return
				}
			}
			http.Error(w, `{"error":"forbidden"}`, http.StatusForbidden)
		})
	}
}
