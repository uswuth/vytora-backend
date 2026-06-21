package vendor

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository struct {
	pool *pgxpool.Pool
}

func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{pool: pool}
}

func (r *Repository) Create(ctx context.Context, vendor *Vendor) error {
	return r.pool.QueryRow(ctx, `
		INSERT INTO vendors (code, name, category, contact_person, contact_email, country,
			contract_start_date, contract_end_date, risk_level, status, assigned_dept_manager_id, created_by)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
		RETURNING id, created_at, updated_at
	`, vendor.Code, vendor.Name, vendor.Category, vendor.ContactPerson, vendor.ContactEmail,
		vendor.Country, vendor.ContractStartDate, vendor.ContractEndDate, vendor.RiskLevel,
		vendor.Status, vendor.AssignedDeptManagerID, vendor.CreatedBy).
		Scan(&vendor.ID, &vendor.CreatedAt, &vendor.UpdatedAt)
}

func (r *Repository) FindByCode(ctx context.Context, code string) (*Vendor, error) {
	vendor := &Vendor{}
	err := r.pool.QueryRow(ctx, `
		SELECT id, code, name, category, contact_person, contact_email, country,
			contract_start_date, contract_end_date, risk_level, status, assigned_dept_manager_id,
			created_by, created_at, updated_at
		FROM vendors
		WHERE code = $1
	`, code).Scan(
		&vendor.ID, &vendor.Code, &vendor.Name, &vendor.Category,
		&vendor.ContactPerson, &vendor.ContactEmail, &vendor.Country,
		&vendor.ContractStartDate, &vendor.ContractEndDate, &vendor.RiskLevel,
		&vendor.Status, &vendor.AssignedDeptManagerID, &vendor.CreatedBy,
		&vendor.CreatedAt, &vendor.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return vendor, nil
}

type ListParams struct {
	Search    string
	Category  string
	RiskLevel string
	Status    string
	Country   string
	CreatedBy *uuid.UUID
	Limit     int
	Offset    int
	SortBy    string
	SortOrder string
}

func (r *Repository) List(ctx context.Context, params ListParams) ([]Vendor, int, error) {
	whereClauses := []string{}
	args := []interface{}{}
	argIdx := 1

	if params.Search != "" {
		whereClauses = append(whereClauses, fmt.Sprintf("(name ILIKE $%d OR code ILIKE $%d)", argIdx, argIdx+1))
		searchPattern := "%" + params.Search + "%"
		args = append(args, searchPattern, searchPattern)
		argIdx += 2
	}
	if params.Category != "" {
		whereClauses = append(whereClauses, fmt.Sprintf("category = $%d", argIdx))
		args = append(args, params.Category)
		argIdx++
	}
	if params.RiskLevel != "" {
		whereClauses = append(whereClauses, fmt.Sprintf("risk_level = $%d", argIdx))
		args = append(args, params.RiskLevel)
		argIdx++
	}
	if params.Status != "" {
		whereClauses = append(whereClauses, fmt.Sprintf("status = $%d", argIdx))
		args = append(args, params.Status)
		argIdx++
	}
	if params.Country != "" {
		whereClauses = append(whereClauses, fmt.Sprintf("country = $%d", argIdx))
		args = append(args, params.Country)
		argIdx++
	}
	if params.CreatedBy != nil {
		whereClauses = append(whereClauses, fmt.Sprintf("created_by = $%d", argIdx))
		args = append(args, *params.CreatedBy)
		argIdx++
	}

	whereSQL := ""
	if len(whereClauses) > 0 {
		whereSQL = "WHERE " + strings.Join(whereClauses, " AND ")
	}

	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM vendors %s", whereSQL)
	var total int
	err := r.pool.QueryRow(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	sortBy := "created_at"
	if params.SortBy != "" {
		sortBy = params.SortBy
	}
	sortOrder := "DESC"
	if strings.ToUpper(params.SortOrder) == "ASC" {
		sortOrder = "ASC"
	}

	limit := 20
	if params.Limit > 0 {
		limit = params.Limit
	}
	offset := 0
	if params.Offset > 0 {
		offset = params.Offset
	}

	query := fmt.Sprintf(`
		SELECT id, code, name, category, contact_person, contact_email, country,
			contract_start_date, contract_end_date, risk_level, status, assigned_dept_manager_id,
			created_by, created_at, updated_at
		FROM vendors
		%s
		ORDER BY %s %s
		LIMIT $%d OFFSET $%d
	`, whereSQL, sortBy, sortOrder, argIdx, argIdx+1)

	args = append(args, limit, offset)

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var vendors []Vendor
	for rows.Next() {
		var v Vendor
		if err := rows.Scan(
			&v.ID, &v.Code, &v.Name, &v.Category,
			&v.ContactPerson, &v.ContactEmail, &v.Country,
			&v.ContractStartDate, &v.ContractEndDate, &v.RiskLevel,
			&v.Status, &v.AssignedDeptManagerID, &v.CreatedBy,
			&v.CreatedAt, &v.UpdatedAt,
		); err != nil {
			return nil, 0, err
		}
		vendors = append(vendors, v)
	}
	return vendors, total, nil
}

func (r *Repository) Update(ctx context.Context, vendor *Vendor) error {
	result, err := r.pool.Exec(ctx, `
		UPDATE vendors
		SET name=$1, category=$2, contact_person=$3, contact_email=$4, country=$5,
			contract_start_date=$6, contract_end_date=$7, risk_level=$8, status=$9,
			assigned_dept_manager_id=$10, updated_at=NOW()
		WHERE code=$11
	`, vendor.Name, vendor.Category, vendor.ContactPerson, vendor.ContactEmail, vendor.Country,
		vendor.ContractStartDate, vendor.ContractEndDate, vendor.RiskLevel, vendor.Status,
		vendor.AssignedDeptManagerID, vendor.Code)
	if err != nil {
		return err
	}
	if result.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}
	return nil
}

func (r *Repository) UpdateStatus(ctx context.Context, code, status string) error {
	_, err := r.pool.Exec(ctx, `UPDATE vendors SET status=$1, updated_at=NOW() WHERE code=$2`, status, code)
	return err
}

func (r *Repository) Delete(ctx context.Context, code string) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM vendors WHERE code = $1`, code)
	return err
}

func (r *Repository) IsActive(ctx context.Context, name string) (bool, error) {
	var status string
	err := r.pool.QueryRow(ctx, `SELECT status FROM vendors WHERE LOWER(name) = LOWER($1)`, name).Scan(&status)
	if err != nil {
		if err == pgx.ErrNoRows {
			return false, nil
		}
		return false, err
	}
	return status == "Active", nil
}