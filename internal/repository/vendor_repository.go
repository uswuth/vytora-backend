package repository

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/uswuth/vytora-backend/internal/models"
)

type VendorRepository struct {
	pool *pgxpool.Pool
}

func NewVendorRepository(pool *pgxpool.Pool) *VendorRepository {
	return &VendorRepository{pool: pool}
}

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
