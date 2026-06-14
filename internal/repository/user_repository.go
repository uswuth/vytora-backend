package repository

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/uswuth/vytora-backend/internal/models"
)

type UserRepository struct {
	pool *pgxpool.Pool
}

func NewUserRepository(pool *pgxpool.Pool) *UserRepository {
	return &UserRepository{pool: pool}
}

func (r *UserRepository) FindByEmail(ctx context.Context, email string) (*models.User, error) {
	user := &models.User{}
	err := r.pool.QueryRow(ctx, `
	SELECT id, code, email, password_hash, full_name, role, is_active, created_at, updated_at
	FROM users
	WHERE email = $1
	`, email).Scan(
		&user.ID, &user.Code, &user.Email, &user.PasswordHash,
		&user.FullName, &user.Role, &user.IsActive,
		&user.CreatedAt, &user.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return user, nil
}
