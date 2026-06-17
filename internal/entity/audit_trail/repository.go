package audit_trail

import (
	"context"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository struct {
	pool *pgxpool.Pool
}

func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{pool: pool}
}

type ListParams struct {
	TableName  string
	RecordCode string // search by entity code (e.g., VEN005)
	Action     string // CREATE, UPDATE, DELETE
	ChangedBy  string // user code (e.g., USR001)
	DateFrom   string // YYYY-MM-DD
	DateTo     string
	Limit      int
	Offset     int
}

func (r *Repository) Create(ctx context.Context, audit *AuditTrail) error {
	return r.pool.QueryRow(ctx, `
		INSERT INTO audit_trail (code, table_name, record_id, action, field_name, old_value, new_value, changed_by)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8)
		RETURNING id, changed_at
	`, audit.Code, audit.TableName, audit.RecordID, audit.Action,
		audit.FieldName, audit.OldValue, audit.NewValue, audit.ChangedBy).
		Scan(&audit.ID, &audit.ChangedAt)
}

func (r *Repository) List(ctx context.Context, params ListParams) ([]AuditTrail, int, error) {
	where := []string{}
	args := []interface{}{}
	idx := 1

	if params.TableName != "" {
		where = append(where, fmt.Sprintf("a.table_name = $%d", idx))
		args = append(args, params.TableName)
		idx++
	}
	if params.RecordCode != "" {
		where = append(where, fmt.Sprintf("(v.code = $%d OR ra.code = $%d OR cr.code = $%d OR c.code = $%d)", idx, idx+1, idx+2, idx+3))
		args = append(args, params.RecordCode, params.RecordCode, params.RecordCode, params.RecordCode)
		idx += 4
	}
	if params.Action != "" {
		where = append(where, fmt.Sprintf("a.action = $%d", idx))
		args = append(args, params.Action)
		idx++
	}
	if params.ChangedBy != "" {
		where = append(where, fmt.Sprintf("u.code = $%d", idx))
		args = append(args, params.ChangedBy)
		idx++
	}
	if params.DateFrom != "" {
		where = append(where, fmt.Sprintf("a.changed_at >= $%d", idx))
		args = append(args, params.DateFrom)
		idx++
	}
	if params.DateTo != "" {
		where = append(where, fmt.Sprintf("a.changed_at <= $%d", idx))
		args = append(args, params.DateTo)
		idx++
	}

	whereSQL := ""
	if len(where) > 0 {
		whereSQL = "WHERE " + strings.Join(where, " AND ")
	}

	// count
	var total int
	countSQL := fmt.Sprintf(`
		SELECT COUNT(*) FROM audit_trail a
		LEFT JOIN users u ON a.changed_by = u.id
		LEFT JOIN vendors v ON a.table_name = 'vendors' AND a.record_id = v.id
		LEFT JOIN risk_assessments ra ON a.table_name = 'risk_assessments' AND a.record_id = ra.id
		LEFT JOIN compliance_records cr ON a.table_name = 'compliance_records' AND a.record_id = cr.id
		LEFT JOIN contracts c ON a.table_name = 'contracts' AND a.record_id = c.id
		%s
	`, whereSQL)
	err := r.pool.QueryRow(ctx, countSQL, args...).Scan(&total)
	if err != nil {
		return nil, 0, err
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
    SELECT a.id, a.code, a.table_name, a.record_id,
        COALESCE(v.code, ra.code, cr.code, c.code) as record_code,
        a.action, a.field_name, a.old_value, a.new_value, a.changed_by, a.changed_at
    FROM audit_trail a
    LEFT JOIN users u ON a.changed_by = u.id
    LEFT JOIN vendors v ON a.table_name = 'vendors' AND a.record_id = v.id
    LEFT JOIN risk_assessments ra ON a.table_name = 'risk_assessments' AND a.record_id = ra.id
    LEFT JOIN compliance_records cr ON a.table_name = 'compliance_records' AND a.record_id = cr.id
    LEFT JOIN contracts c ON a.table_name = 'contracts' AND a.record_id = c.id
    %s
    ORDER BY a.changed_at DESC
    LIMIT $%d OFFSET $%d
`, whereSQL, idx, idx+1)
	args = append(args, limit, offset)

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var audits []AuditTrail
	for rows.Next() {
		var a AuditTrail
		if err := rows.Scan(
			&a.ID, &a.Code, &a.TableName, &a.RecordID, &a.RecordCode,
			&a.Action, &a.FieldName, &a.OldValue, &a.NewValue,
			&a.ChangedBy, &a.ChangedAt,
		); err != nil {
			return nil, 0, err
		}
		audits = append(audits, a)
	}
	return audits, total, nil
}