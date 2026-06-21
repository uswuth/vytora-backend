package directives

import (
	"context"

	"github.com/99designs/gqlgen/graphql"
	"github.com/uswuth/vytora-backend/internal/middleware"
	graphqlmiddleware "github.com/uswuth/vytora-backend/internal/middleware/graphql"
)

// IsAuthenticated directive ensures that a valid JWT is present.
func IsAuthenticated(ctx context.Context, obj any, next graphql.Resolver) (any, error) {
	_, ok := graphqlmiddleware.GetClaims(ctx)
	if !ok {
		return nil, &UnauthorizedError{}
	}
	return next(ctx)
}

// HasPermission checks whether the authenticated user has the required permission.
func HasPermission(ctx context.Context, obj any, next graphql.Resolver, permission string) (any, error) {
	claims, ok := graphqlmiddleware.GetClaims(ctx)
	if !ok {
		return nil, &UnauthorizedError{}
	}

	allowed, exists := middleware.RolePermissions[claims.Role]
	if !exists {
		return nil, &ForbiddenError{Message: "role has no permissions defined"}
	}

	for _, p := range allowed {
		if p == permission {
			return next(ctx)
		}
	}
	return nil, &ForbiddenError{Message: "insufficient permissions"}
}

// UnauthorizedError represents an unauthenticated request.
type UnauthorizedError struct{}

func (e *UnauthorizedError) Error() string { return "unauthorized" }

// ForbiddenError represents an authorized request with insufficient permissions.
type ForbiddenError struct {
	Message string
}

func (e *ForbiddenError) Error() string { return e.Message }