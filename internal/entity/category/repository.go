package category

import (
	"context"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository struct {
	pool *pgxpool.Pool
}

func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{pool: pool}
}

func (r *Repository) Create(ctx context.Context, c *Category) error {
	return r.pool.QueryRow(ctx, `
		INSERT INTO categories (code, name, display_name, description, status, created_by)
		VALUES ($1,$2,$3,$4,$5,$6)
		RETURNING id, created_at, updated_at
	`, c.Code, c.Name, c.DisplayName, c.Description, c.Status, c.CreatedBy).
		Scan(&c.ID, &c.CreatedAt, &c.UpdatedAt)
}

func (r *Repository) FindByCode(ctx context.Context, code string) (*Category, error) {
	c := &Category{}
	err := r.pool.QueryRow(ctx, `
		SELECT id, code, name, display_name, description, status, created_by, updated_by, created_at, updated_at
		FROM categories WHERE code = $1
	`, code).Scan(&c.ID, &c.Code, &c.Name, &c.DisplayName, &c.Description, &c.Status,
		&c.CreatedBy, &c.UpdatedBy, &c.CreatedAt, &c.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return c, nil
}

func (r *Repository) FindByName(ctx context.Context, name string) (*Category, error) {
	c := &Category{}
	err := r.pool.QueryRow(ctx, `
		SELECT id, code, name, display_name, description, status, created_by, updated_by, created_at, updated_at
		FROM categories WHERE LOWER(name) = LOWER($1)
	`, name).Scan(&c.ID, &c.Code, &c.Name, &c.DisplayName, &c.Description, &c.Status,
		&c.CreatedBy, &c.UpdatedBy, &c.CreatedAt, &c.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return c, nil
}

type ListParams struct {
	Search string
	Status string
	Limit  int
	Offset int
}

func (r *Repository) List(ctx context.Context, params ListParams) ([]Category, int, error) {
	where := []string{}
	args := []interface{}{}
	idx := 1

	if params.Search != "" {
		where = append(where, fmt.Sprintf("(name ILIKE $%d OR display_name ILIKE $%d)", idx, idx+1))
		args = append(args, "%"+params.Search+"%", "%"+params.Search+"%")
		idx += 2
	}
	if params.Status != "" {
		where = append(where, fmt.Sprintf("status = $%d", idx))
		args = append(args, params.Status)
		idx++
	}

	whereSQL := ""
	if len(where) > 0 {
		whereSQL = "WHERE " + strings.Join(where, " AND ")
	}

	countSQL := fmt.Sprintf("SELECT COUNT(*) FROM categories %s", whereSQL)
	var total int
	err := r.pool.QueryRow(ctx, countSQL, args...).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	limit := 50
	if params.Limit > 0 {
		limit = params.Limit
	}
	offset := 0
	if params.Offset > 0 {
		offset = params.Offset
	}

	query := fmt.Sprintf(`
		SELECT id, code, name, display_name, description, status, created_by, updated_by, created_at, updated_at
		FROM categories
		%s
		ORDER BY name ASC
		LIMIT $%d OFFSET $%d
	`, whereSQL, idx, idx+1)
	args = append(args, limit, offset)

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var cats []Category
	for rows.Next() {
		var c Category
		if err := rows.Scan(&c.ID, &c.Code, &c.Name, &c.DisplayName, &c.Description, &c.Status,
			&c.CreatedBy, &c.UpdatedBy, &c.CreatedAt, &c.UpdatedAt); err != nil {
			return nil, 0, err
		}
		cats = append(cats, c)
	}
	return cats, total, nil
}

func (r *Repository) Update(ctx context.Context, c *Category) error {
	_, err := r.pool.Exec(ctx, `
		UPDATE categories SET display_name=$1, description=$2, status=$3, updated_by=$4, updated_at=NOW()
		WHERE code=$5
	`, c.DisplayName, c.Description, c.Status, c.UpdatedBy, c.Code)
	return err
}

func (r *Repository) Delete(ctx context.Context, code string) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM categories WHERE code=$1`, code)
	return err
}

func (r *Repository) IsActive(ctx context.Context, name string) (bool, error) {
	var status string
	err := r.pool.QueryRow(ctx, `SELECT status FROM categories WHERE LOWER(name) = LOWER($1)`, name).Scan(&status)
	if err != nil {
		if err == pgx.ErrNoRows {
			return false, nil
		}
		return false, err
	}
	return status == "Active", nil
}