package contract

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository struct {
	pool *pgxpool.Pool
}

func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{pool: pool}
}

func (r *Repository) Create(ctx context.Context, c *Contract) error {
	return r.pool.QueryRow(ctx, `
		INSERT INTO contracts (code, vendor_id, contract_number, start_date, end_date,
			contract_value, renewal_status)
		VALUES ($1,$2,$3,$4,$5,$6,$7)
		RETURNING id, created_at, updated_at
	`, c.Code, c.VendorID, c.ContractNumber, c.StartDate, c.EndDate,
		c.ContractValue, c.RenewalStatus).
		Scan(&c.ID, &c.CreatedAt, &c.UpdatedAt)
}

func (r *Repository) FindByCode(ctx context.Context, code string) (*Contract, error) {
	c := &Contract{}
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

func (r *Repository) ListByVendor(ctx context.Context, vendorID string) ([]Contract, error) {
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

	var contracts []Contract
	for rows.Next() {
		var c Contract
		if err := rows.Scan(&c.ID, &c.Code, &c.VendorID, &c.VendorCode, &c.ContractNumber,
			&c.StartDate, &c.EndDate, &c.ContractValue, &c.RenewalStatus,
			&c.CreatedAt, &c.UpdatedAt); err != nil {
			return nil, err
		}
		contracts = append(contracts, c)
	}
	return contracts, nil
}

func (r *Repository) Update(ctx context.Context, c *Contract) error {
	result, err := r.pool.Exec(ctx, `
		UPDATE contracts
		SET vendor_id=$1, contract_number=$2, start_date=$3, end_date=$4,
		    contract_value=$5, renewal_status=$6, updated_at=NOW()
		WHERE code=$7
	`, c.VendorID, c.ContractNumber, c.StartDate, c.EndDate,
		c.ContractValue, c.RenewalStatus, c.Code)
	if err != nil {
		return err
	}
	if result.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}
	return nil
}

func (r *Repository) Delete(ctx context.Context, code string) error {
	result, err := r.pool.Exec(ctx, `DELETE FROM contracts WHERE code = $1`, code)
	if err != nil {
		return err
	}
	if result.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}
	return nil
}

func (r *Repository) Expiring(ctx context.Context, days int) ([]Contract, error) {
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

	var contracts []Contract
	for rows.Next() {
		var c Contract
		if err := rows.Scan(&c.ID, &c.Code, &c.VendorID, &c.VendorCode, &c.ContractNumber,
			&c.StartDate, &c.EndDate, &c.ContractValue, &c.RenewalStatus,
			&c.CreatedAt, &c.UpdatedAt); err != nil {
			return nil, err
		}
		contracts = append(contracts, c)
	}
	return contracts, nil
}