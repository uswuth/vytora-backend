package repository

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/uswuth/vytora-backend/internal/models"
)

type AuditTrailRepository struct {
	pool *pgxpool.Pool
}

func NewAuditTrailRepository(pool *pgxpool.Pool) *AuditTrailRepository {
	return &AuditTrailRepository{pool: pool}
}

func (r *AuditTrailRepository) Create(ctx context.Context, audit *models.AuditTrail) error {
	return r.pool.QueryRow(ctx, `
		INSERT INTO audit_trail (code, table_name, record_id, action, field_name, old_value, new_value, changed_by)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8)
		RETURNING id, changed_at
	`, audit.Code, audit.TableName, audit.RecordID, audit.Action,
		audit.FieldName, audit.OldValue, audit.NewValue, audit.ChangedBy).
		Scan(&audit.ID, &audit.ChangedAt)
}