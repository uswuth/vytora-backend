package compliance_record

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository struct {
	pool *pgxpool.Pool
}

func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{pool: pool}
}

func (r *Repository) Create(ctx context.Context, cr *ComplianceRecord) error {
	return r.pool.QueryRow(ctx, `
		INSERT INTO compliance_records (code, vendor_id, certification_type, status,
			valid_from, valid_until, issued_by, evidence_url, reviewed_by)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9)
		RETURNING id, created_at, updated_at
	`, cr.Code, cr.VendorID, cr.CertificationType, cr.Status,
		cr.ValidFrom, cr.ValidUntil, cr.IssuedBy, cr.EvidenceURL, cr.ReviewedBy).
		Scan(&cr.ID, &cr.CreatedAt, &cr.UpdatedAt)
}

func (r *Repository) FindByCode(ctx context.Context, code string) (*ComplianceRecord, error) {
	cr := &ComplianceRecord{}
	err := r.pool.QueryRow(ctx, `
		SELECT cr.id, cr.code, cr.vendor_id, v.code, cr.certification_type, cr.status,
			cr.valid_from, cr.valid_until, cr.issued_by, cr.evidence_url, cr.reviewed_by,
			cr.created_at, cr.updated_at
		FROM compliance_records cr
		JOIN vendors v ON cr.vendor_id = v.id
		WHERE cr.code = $1
	`, code).Scan(
		&cr.ID, &cr.Code, &cr.VendorID, &cr.VendorCode, &cr.CertificationType, &cr.Status,
		&cr.ValidFrom, &cr.ValidUntil, &cr.IssuedBy, &cr.EvidenceURL, &cr.ReviewedBy,
		&cr.CreatedAt, &cr.UpdatedAt,
	)
	return cr, err
}

func (r *Repository) Update(ctx context.Context, cr *ComplianceRecord) error {
	_, err := r.pool.Exec(ctx, `
		UPDATE compliance_records SET certification_type=$1, status=$2, valid_from=$3,
			valid_until=$4, issued_by=$5, evidence_url=$6, reviewed_by=$7, updated_at=NOW()
		WHERE code=$8
	`, cr.CertificationType, cr.Status, cr.ValidFrom, cr.ValidUntil,
		cr.IssuedBy, cr.EvidenceURL, cr.ReviewedBy, cr.Code)
	return err
}

func (r *Repository) ListByVendor(ctx context.Context, vendorID string) ([]ComplianceRecord, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT cr.id, cr.code, cr.vendor_id, v.code, cr.certification_type, cr.status,
			cr.valid_from, cr.valid_until, cr.issued_by, cr.evidence_url, cr.reviewed_by,
			cr.created_at, cr.updated_at
		FROM compliance_records cr
		JOIN vendors v ON cr.vendor_id = v.id
		WHERE cr.vendor_id = $1
		ORDER BY cr.created_at DESC
	`, vendorID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var records []ComplianceRecord
	for rows.Next() {
		var cr ComplianceRecord
		if err := rows.Scan(&cr.ID, &cr.Code, &cr.VendorID, &cr.VendorCode,
			&cr.CertificationType, &cr.Status, &cr.ValidFrom, &cr.ValidUntil,
			&cr.IssuedBy, &cr.EvidenceURL, &cr.ReviewedBy,
			&cr.CreatedAt, &cr.UpdatedAt); err != nil {
			return nil, err
		}
		records = append(records, cr)
	}
	return records, nil
}

// Expiring returns certifications expiring within the given number of days.
func (r *Repository) Expiring(ctx context.Context, days int) ([]ComplianceRecord, error) {
	cutoff := time.Now().AddDate(0, 0, days)
	rows, err := r.pool.Query(ctx, `
		SELECT cr.id, cr.code, cr.vendor_id, v.code, cr.certification_type, cr.status,
			cr.valid_from, cr.valid_until, cr.issued_by, cr.evidence_url, cr.reviewed_by,
			cr.created_at, cr.updated_at
		FROM compliance_records cr
		JOIN vendors v ON cr.vendor_id = v.id
		WHERE cr.status = 'Approved' AND cr.valid_until <= $1
		ORDER BY cr.valid_until ASC
	`, cutoff)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var records []ComplianceRecord
	for rows.Next() {
		var cr ComplianceRecord
		if err := rows.Scan(&cr.ID, &cr.Code, &cr.VendorID, &cr.VendorCode,
			&cr.CertificationType, &cr.Status, &cr.ValidFrom, &cr.ValidUntil,
			&cr.IssuedBy, &cr.EvidenceURL, &cr.ReviewedBy,
			&cr.CreatedAt, &cr.UpdatedAt); err != nil {
			return nil, err
		}
		records = append(records, cr)
	}
	return records, nil
}