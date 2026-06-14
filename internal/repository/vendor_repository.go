package repository

import (
	"context"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/uswuth/vytora-backend/internal/models"
)

type VendorRepository struct {
	pool *pgxpool.Pool
}

func NewVendorRepository(pool *pgxpool.Pool) *VendorRepository {
	return &VendorRepository{pool: pool}
}

// Create inserts a new vendor and populates ID, CreatedAt, UpdatedAt.
func (r *VendorRepository) Create(ctx context.Context, vendor *models.Vendor) error {
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

// FindByCode retrieves a vendor by its unique code (e.g., VEN005).
func (r *VendorRepository) FindByCode(ctx context.Context, code string) (*models.Vendor, error) {
	vendor := &models.Vendor{}
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

// ListParams holds the optional filters, search, and pagination.
type ListParams struct {
	Search    string // search in name or code
	Category  string
	RiskLevel string
	Status    string
	Country   string
	Limit     int
	Offset    int
	SortBy    string // column to sort by (default: created_at)
	SortOrder string // ASC or DESC (default: DESC)
}

// List retrieves vendors with filtering, search, pagination, and sorting.
func (r *VendorRepository) List(ctx context.Context, params ListParams) ([]models.Vendor, int, error) {
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

	whereSQL := ""
	if len(whereClauses) > 0 {
		whereSQL = "WHERE " + strings.Join(whereClauses, " AND ")
	}

	// Count total matching records
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM vendors %s", whereSQL)
	var total int
	err := r.pool.QueryRow(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	// Sorting
	sortBy := "created_at"
	if params.SortBy != "" {
		sortBy = params.SortBy
	}
	sortOrder := "DESC"
	if strings.ToUpper(params.SortOrder) == "ASC" {
		sortOrder = "ASC"
	}

	// Pagination
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

	var vendors []models.Vendor
	for rows.Next() {
		var v models.Vendor
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

// Update modifies an existing vendor. Returns error if not found.
func (r *VendorRepository) Update(ctx context.Context, vendor *models.Vendor) error {
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

// UpdateStatus changes only the status field (used by workflow).
func (r *VendorRepository) UpdateStatus(ctx context.Context, code, status string) error {
	_, err := r.pool.Exec(ctx, `UPDATE vendors SET status=$1, updated_at=NOW() WHERE code=$2`, status, code)
	return err
}

// Delete removes a vendor by code.
func (r *VendorRepository) Delete(ctx context.Context, code string) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM vendors WHERE code = $1`, code)
	return err
}
