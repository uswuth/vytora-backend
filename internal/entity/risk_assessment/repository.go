package risk_assessment

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

func (r *Repository) Create(ctx context.Context, ra *RiskAssessment) error {
	return r.pool.QueryRow(ctx, `
		INSERT INTO risk_assessments (code, vendor_id, vendor_code, assessment_date, risk_level, findings, recommendations, status, reviewed_by, reviewed_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)
		RETURNING id, created_at, updated_at
	`, ra.Code, ra.VendorID, ra.VendorCode, ra.AssessmentDate, ra.RiskLevel, ra.Findings, ra.Recommendations, ra.Status, ra.ReviewedBy, ra.ReviewedAt).
		Scan(&ra.ID, &ra.CreatedAt, &ra.UpdatedAt)
}

func (r *Repository) FindByCode(ctx context.Context, code string) (*RiskAssessment, error) {
	ra := &RiskAssessment{}
	err := r.pool.QueryRow(ctx, `
		SELECT id, code, vendor_id, vendor_code, assessment_date, risk_level, findings, recommendations, status, reviewed_by, reviewed_at, created_at, updated_at
		FROM risk_assessments
		WHERE code = $1
	`, code).Scan(
		&ra.ID, &ra.Code, &ra.VendorID, &ra.VendorCode, &ra.AssessmentDate, &ra.RiskLevel, &ra.Findings, &ra.Recommendations, &ra.Status, &ra.ReviewedBy, &ra.ReviewedAt, &ra.CreatedAt, &ra.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return ra, nil
}

func (r *Repository) List(ctx context.Context, vendorID string) ([]RiskAssessment, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, code, vendor_id, vendor_code, assessment_date, risk_level, findings, recommendations, status, reviewed_by, reviewed_at, created_at, updated_at
		FROM risk_assessments
		WHERE vendor_id = $1
		ORDER BY assessment_date DESC
	`, vendorID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var assessments []RiskAssessment
	for rows.Next() {
		var ra RiskAssessment
		if err := rows.Scan(
			&ra.ID, &ra.Code, &ra.VendorID, &ra.VendorCode, &ra.AssessmentDate, &ra.RiskLevel, &ra.Findings, &ra.Recommendations, &ra.Status, &ra.ReviewedBy, &ra.ReviewedAt, &ra.CreatedAt, &ra.UpdatedAt,
		); err != nil {
			return nil, err
		}
		assessments = append(assessments, ra)
	}
	return assessments, nil
}

func (r *Repository) Update(ctx context.Context, ra *RiskAssessment) error {
	result, err := r.pool.Exec(ctx, `
		UPDATE risk_assessments
		SET risk_level=$1, findings=$2, recommendations=$3, status=$4, reviewed_by=$5, reviewed_at=$6, updated_at=NOW()
		WHERE code=$7
	`, ra.RiskLevel, ra.Findings, ra.Recommendations, ra.Status, ra.ReviewedBy, ra.ReviewedAt, ra.Code)
	if err != nil {
		return err
	}
	if result.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}
	return nil
}

type ListParams struct {
	VendorCode string
	RiskLevel  string
	Status     string
	Limit      int
	Offset     int
}

func (r *Repository) ListAll(ctx context.Context, params ListParams) ([]RiskAssessment, int, error) {
	where := []string{}
	args := []interface{}{}
	idx := 1

	if params.VendorCode != "" {
		where = append(where, fmt.Sprintf("vendor_code = $%d", idx))
		args = append(args, params.VendorCode)
		idx++
	}
	if params.RiskLevel != "" {
		where = append(where, fmt.Sprintf("risk_level = $%d", idx))
		args = append(args, params.RiskLevel)
		idx++
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

	countSQL := fmt.Sprintf("SELECT COUNT(*) FROM risk_assessments %s", whereSQL)
	var total int
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
		SELECT id, code, vendor_id, vendor_code, assessment_date, risk_level, findings, recommendations, status, reviewed_by, reviewed_at, created_at, updated_at
		FROM risk_assessments
		%s
		ORDER BY assessment_date DESC
		LIMIT $%d OFFSET $%d
	`, whereSQL, idx, idx+1)
	args = append(args, limit, offset)

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var assessments []RiskAssessment
	for rows.Next() {
		var ra RiskAssessment
		if err := rows.Scan(
			&ra.ID, &ra.Code, &ra.VendorID, &ra.VendorCode, &ra.AssessmentDate, &ra.RiskLevel, &ra.Findings, &ra.Recommendations, &ra.Status, &ra.ReviewedBy, &ra.ReviewedAt, &ra.CreatedAt, &ra.UpdatedAt,
		); err != nil {
			return nil, 0, err
		}
		assessments = append(assessments, ra)
	}
	return assessments, total, nil
}