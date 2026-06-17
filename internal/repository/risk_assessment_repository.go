package repository

import (
	"context"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/uswuth/vytora-backend/internal/models"
)

type RiskAssessmentRepository struct {
	pool *pgxpool.Pool
}

func NewRiskAssessmentRepository(pool *pgxpool.Pool) *RiskAssessmentRepository {
	return &RiskAssessmentRepository{pool: pool}
}

func (r *RiskAssessmentRepository) Create(ctx context.Context, ra *models.RiskAssessment) error {
	return r.pool.QueryRow(ctx, `
		INSERT INTO risk_assessments (code, vendor_id, assessment_date, assessor_id,
			overall_risk_score, risk_level, security_risk_score, financial_risk_score,
			operational_risk_score, legal_risk_score, status, notes)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12)
		RETURNING id, created_at, updated_at
	`, ra.Code, ra.VendorID, ra.AssessmentDate, ra.AssessorID,
		ra.OverallRiskScore, ra.RiskLevel, ra.SecurityRiskScore, ra.FinancialRiskScore,
		ra.OperationalRiskScore, ra.LegalRiskScore, ra.Status, ra.Notes).
		Scan(&ra.ID, &ra.CreatedAt, &ra.UpdatedAt)
}

func (r *RiskAssessmentRepository) FindByCode(ctx context.Context, code string) (*models.RiskAssessment, error) {
	ra := &models.RiskAssessment{}
	err := r.pool.QueryRow(ctx, `
		SELECT ra.id, ra.code, ra.vendor_id, v.code, ra.assessment_date, ra.assessor_id, u.code,
			ra.overall_risk_score, ra.risk_level, ra.security_risk_score, ra.financial_risk_score,
			ra.operational_risk_score, ra.legal_risk_score, ra.status, ra.notes,
			ra.created_at, ra.updated_at
		FROM risk_assessments ra
		JOIN vendors v ON ra.vendor_id = v.id
		JOIN users u ON ra.assessor_id = u.id
		WHERE ra.code = $1
	`, code).Scan(
		&ra.ID, &ra.Code, &ra.VendorID, &ra.VendorCode, &ra.AssessmentDate,
		&ra.AssessorID, &ra.AssessorCode,
		&ra.OverallRiskScore, &ra.RiskLevel, &ra.SecurityRiskScore, &ra.FinancialRiskScore,
		&ra.OperationalRiskScore, &ra.LegalRiskScore, &ra.Status, &ra.Notes,
		&ra.CreatedAt, &ra.UpdatedAt,
	)
	return ra, err
}

type RiskListParams struct {
	VendorCode string
	RiskLevel  string
	Status     string
	DateFrom   string
	DateTo     string
	Limit      int
	Offset     int
}

func (r *RiskAssessmentRepository) List(ctx context.Context, params RiskListParams) ([]models.RiskAssessment, int, error) {
	where := []string{}
	args := []interface{}{}
	idx := 1

	if params.VendorCode != "" {
		where = append(where, fmt.Sprintf("v.code = $%d", idx))
		args = append(args, params.VendorCode)
		idx++
	}
	if params.RiskLevel != "" {
		where = append(where, fmt.Sprintf("ra.risk_level = $%d", idx))
		args = append(args, params.RiskLevel)
		idx++
	}
	if params.Status != "" {
		where = append(where, fmt.Sprintf("ra.status = $%d", idx))
		args = append(args, params.Status)
		idx++
	}
	if params.DateFrom != "" {
		where = append(where, fmt.Sprintf("ra.assessment_date >= $%d", idx))
		args = append(args, params.DateFrom)
		idx++
	}
	if params.DateTo != "" {
		where = append(where, fmt.Sprintf("ra.assessment_date <= $%d", idx))
		args = append(args, params.DateTo)
		idx++
	}

	whereSQL := ""
	if len(where) > 0 {
		whereSQL = "WHERE " + strings.Join(where, " AND ")
	}

	// total
	countSQL := fmt.Sprintf(`
		SELECT COUNT(*) FROM risk_assessments ra
		JOIN vendors v ON ra.vendor_id = v.id
		JOIN users u ON ra.assessor_id = u.id
		%s
	`, whereSQL)
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
		SELECT ra.id, ra.code, ra.vendor_id, v.code, ra.assessment_date, ra.assessor_id, u.code,
			ra.overall_risk_score, ra.risk_level, ra.security_risk_score,
			ra.financial_risk_score, ra.operational_risk_score, ra.legal_risk_score,
			ra.status, ra.notes, ra.created_at, ra.updated_at
		FROM risk_assessments ra
		JOIN vendors v ON ra.vendor_id = v.id
		JOIN users u ON ra.assessor_id = u.id
		%s
		ORDER BY ra.created_at DESC
		LIMIT $%d OFFSET $%d
	`, whereSQL, idx, idx+1)
	args = append(args, limit, offset)

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var assessments []models.RiskAssessment
	for rows.Next() {
		var a models.RiskAssessment
		if err := rows.Scan(
			&a.ID, &a.Code, &a.VendorID, &a.VendorCode, &a.AssessmentDate,
			&a.AssessorID, &a.AssessorCode,
			&a.OverallRiskScore, &a.RiskLevel, &a.SecurityRiskScore,
			&a.FinancialRiskScore, &a.OperationalRiskScore, &a.LegalRiskScore,
			&a.Status, &a.Notes, &a.CreatedAt, &a.UpdatedAt,
		); err != nil {
			return nil, 0, err
		}
		assessments = append(assessments, a)
	}
	return assessments, total, nil
}

func (r *RiskAssessmentRepository) UpdateStatus(ctx context.Context, code, status string) error {
	_, err := r.pool.Exec(ctx, `UPDATE risk_assessments SET status=$1, updated_at=NOW() WHERE code=$2`, status, code)
	return err
}