package repository

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/uswuth/vytora-backend/internal/models"
)

type ContractRepository struct {
	pool *pgxpool.Pool
}

func NewContractRepository(pool *pgxpool.Pool) *ContractRepository {
	return &ContractRepository{pool: pool}
}

func (r *ContractRepository) Create(ctx context.Context, c *models.Contract) error {
	return r.pool.QueryRow(ctx, `
		INSERT INTO contracts (code, vendor_id, contract_number, start_date, end_date,
			contract_value, renewal_status)
		VALUES ($1,$2,$3,$4,$5,$6,$7)
		RETURNING id, created_at, updated_at
	`, c.Code, c.VendorID, c.ContractNumber, c.StartDate, c.EndDate,
		c.ContractValue, c.RenewalStatus).
		Scan(&c.ID, &c.CreatedAt, &c.UpdatedAt)
}

func (r *ContractRepository) FindByCode(ctx context.Context, code string) (*models.Contract, error) {
	c := &models.Contract{}
	err := r.pool.QueryRow(ctx, `
		SELECT c.id, c.code, c.vendor_id, v.code, c.contract_number, c.start_date,
			c.end_date, c.contract_value, c.renewal_status, c.created_at, c.updated_at
		FROM contracts c
		JOIN vendors v ON c.vendor_id = v.id
		WHERE c.code = $1
	`, code).Scan(
		&c.ID, &c.Code, &c.VendorID, &c.VendorCode, &c.ContractNumber,
		&c.StartDate, &c.EndDate, &c.ContractValue, &c.RenewalStatus,
		&c.CreatedAt, &c.UpdatedAt,
	)
	return c, err
}

func (r *ContractRepository) ListByVendor(ctx context.Context, vendorID string) ([]models.Contract, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT c.id, c.code, c.vendor_id, v.code, c.contract_number, c.start_date,
			c.end_date, c.contract_value, c.renewal_status, c.created_at, c.updated_at
		FROM contracts c
		JOIN vendors v ON c.vendor_id = v.id
		WHERE c.vendor_id = $1
		ORDER BY c.end_date ASC
	`, vendorID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var contracts []models.Contract
	for rows.Next() {
		var c models.Contract
		if err := rows.Scan(&c.ID, &c.Code, &c.VendorID, &c.VendorCode, &c.ContractNumber,
			&c.StartDate, &c.EndDate, &c.ContractValue, &c.RenewalStatus,
			&c.CreatedAt, &c.UpdatedAt); err != nil {
			return nil, err
		}
		contracts = append(contracts, c)
	}
	return contracts, nil
}

func (r *ContractRepository) Expiring(ctx context.Context, days int) ([]models.Contract, error) {
	cutoff := time.Now().AddDate(0, 0, days)
	rows, err := r.pool.Query(ctx, `
		SELECT c.id, c.code, c.vendor_id, v.code, c.contract_number, c.start_date,
			c.end_date, c.contract_value, c.renewal_status, c.created_at, c.updated_at
		FROM contracts c
		JOIN vendors v ON c.vendor_id = v.id
		WHERE c.end_date <= $1
		ORDER BY c.end_date ASC
	`, cutoff)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var contracts []models.Contract
	for rows.Next() {
		var c models.Contract
		if err := rows.Scan(&c.ID, &c.Code, &c.VendorID, &c.VendorCode, &c.ContractNumber,
			&c.StartDate, &c.EndDate, &c.ContractValue, &c.RenewalStatus,
			&c.CreatedAt, &c.UpdatedAt); err != nil {
			return nil, err
		}
		contracts = append(contracts, c)
	}
	return contracts, nil
}