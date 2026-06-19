package user

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository struct {
	pool *pgxpool.Pool
}

func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{pool: pool}
}

func (r *Repository) FindByEmail(ctx context.Context, email string) (*User, error) {
	user := &User{}
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

func (r *Repository) Create(ctx context.Context, user *User) error {
	return r.pool.QueryRow(ctx, `
		INSERT INTO users (code, email, password_hash, full_name, role)
		VALUES ($1,$2,$3,$4,$5)
		RETURNING id, created_at, updated_at
	`, user.Code, user.Email, user.PasswordHash, user.FullName, user.Role).
		Scan(&user.ID, &user.CreatedAt, &user.UpdatedAt)
}

func (r *Repository) FindByID(ctx context.Context, id string) (*User, error) {
	user := &User{}
	err := r.pool.QueryRow(ctx, `
		SELECT id, code, email, full_name, role, is_active, created_at, updated_at
		FROM users WHERE id = $1
	`, id).Scan(&user.ID, &user.Code, &user.Email, &user.FullName,
		&user.Role, &user.IsActive, &user.CreatedAt, &user.UpdatedAt)
	return user, err
}

func (r *Repository) ListAll(ctx context.Context) ([]User, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, code, email, full_name, role, is_active, created_at, updated_at
		FROM users ORDER BY created_at DESC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var users []User
	for rows.Next() {
		var u User
		if err := rows.Scan(&u.ID, &u.Code, &u.Email, &u.FullName,
			&u.Role, &u.IsActive, &u.CreatedAt, &u.UpdatedAt); err != nil {
			return nil, err
		}
		users = append(users, u)
	}
	return users, nil
}

func (r *Repository) UpdateRole(ctx context.Context, id, role string) error {
	_, err := r.pool.Exec(ctx, `UPDATE users SET role=$1, updated_at=NOW() WHERE id=$2`, role, id)
	return err
}

func (r *Repository) SetActive(ctx context.Context, id string, active bool) error {
	_, err := r.pool.Exec(ctx, `UPDATE users SET is_active=$1, updated_at=NOW() WHERE id=$2`, active, id)
	return err
}