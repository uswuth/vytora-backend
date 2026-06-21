package vendor_contact

import (
	"context"
	"fmt"

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

func (r *Repository) Create(ctx context.Context, contact *VendorContact) error {
	return r.pool.QueryRow(ctx, `
		INSERT INTO vendor_contacts (code, vendor_id, name, email, phone)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, created_at
	`, contact.Code, contact.VendorID, contact.Name, contact.Email, contact.Phone).
		Scan(&contact.ID, &contact.CreatedAt)
}

func (r *Repository) FindByID(ctx context.Context, id uuid.UUID) (*VendorContact, error) {
	contact := &VendorContact{}
	err := r.pool.QueryRow(ctx, `
		SELECT id, code, vendor_id, name, email, phone, created_at
		FROM vendor_contacts
		WHERE id = $1
	`, id).Scan(
		&contact.ID, &contact.Code, &contact.VendorID,
		&contact.Name, &contact.Email, &contact.Phone, &contact.CreatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return contact, nil
}

func (r *Repository) ListByVendor(ctx context.Context, vendorID uuid.UUID) ([]VendorContact, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, code, vendor_id, name, email, phone, created_at
		FROM vendor_contacts
		WHERE vendor_id = $1
		ORDER BY created_at DESC
	`, vendorID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var contacts []VendorContact
	for rows.Next() {
		var c VendorContact
		if err := rows.Scan(
			&c.ID, &c.Code, &c.VendorID,
			&c.Name, &c.Email, &c.Phone, &c.CreatedAt,
		); err != nil {
			return nil, err
		}
		contacts = append(contacts, c)
	}
	return contacts, nil
}

func (r *Repository) Update(ctx context.Context, contact *VendorContact) error {
	result, err := r.pool.Exec(ctx, `
		UPDATE vendor_contacts
		SET name=$1, email=$2, phone=$3
		WHERE id=$4
	`, contact.Name, contact.Email, contact.Phone, contact.ID)
	if err != nil {
		return err
	}
	if result.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}
	return nil
}

func (r *Repository) Delete(ctx context.Context, id uuid.UUID) error {
	result, err := r.pool.Exec(ctx, `DELETE FROM vendor_contacts WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("delete vendor contact: %w", err)
	}
	if result.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}
	return nil
}