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
		INSERT INTO risk_assessments
			(code, vendor_id, assessment_date, assessor_id, overall_risk_score,
			 risk_level, security_risk_score, financial_risk_score,
			 operational_risk_score, legal_risk_score, status, notes)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12)
		RETURNING id, created_at, updated_at
	`, ra.Code, ra.VendorID, ra.AssessmentDate, ra.AssessorID,
		ra.OverallRiskScore, ra.RiskLevel, ra.SecurityRiskScore,
		ra.FinancialRiskScore, ra.OperationalRiskScore,
		ra.LegalRiskScore, ra.Status, ra.Notes).
		Scan(&ra.ID, &ra.CreatedAt, &ra.UpdatedAt)
}

func (r *Repository) FindByCode(ctx context.Context, code string) (*RiskAssessment, error) {
	ra := &RiskAssessment{}
	err := r.pool.QueryRow(ctx, `
		SELECT ra.id, ra.code, ra.vendor_id, v.code,
			ra.assessment_date, ra.assessor_id,
			ra.overall_risk_score, ra.risk_level,
			ra.security_risk_score, ra.financial_risk_score,
			ra.operational_risk_score, ra.legal_risk_score,
			ra.status, ra.notes,
			ra.created_at, ra.updated_at
		FROM risk_assessments ra
		JOIN vendors v ON ra.vendor_id = v.id
		WHERE ra.code = $1
	`, code).Scan(
		&ra.ID, &ra.Code, &ra.VendorID, &ra.VendorCode,
		&ra.AssessmentDate, &ra.AssessorID,
		&ra.OverallRiskScore, &ra.RiskLevel,
		&ra.SecurityRiskScore, &ra.FinancialRiskScore,
		&ra.OperationalRiskScore, &ra.LegalRiskScore,
		&ra.Status, &ra.Notes,
		&ra.CreatedAt, &ra.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return ra, nil
}

func (r *Repository) ListByVendor(ctx context.Context, vendorID string) ([]RiskAssessment, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT ra.id, ra.code, ra.vendor_id, v.code,
			ra.assessment_date, ra.assessor_id,
			ra.overall_risk_score, ra.risk_level,
			ra.security_risk_score, ra.financial_risk_score,
			ra.operational_risk_score, ra.legal_risk_score,
			ra.status, ra.notes,
			ra.created_at, ra.updated_at
		FROM risk_assessments ra
		JOIN vendors v ON ra.vendor_id = v.id
		WHERE ra.vendor_id = $1
		ORDER BY ra.assessment_date DESC
	`, vendorID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var assessments []RiskAssessment
	for rows.Next() {
		var ra RiskAssessment
		if err := rows.Scan(
			&ra.ID, &ra.Code, &ra.VendorID, &ra.VendorCode,
			&ra.AssessmentDate, &ra.AssessorID,
			&ra.OverallRiskScore, &ra.RiskLevel,
			&ra.SecurityRiskScore, &ra.FinancialRiskScore,
			&ra.OperationalRiskScore, &ra.LegalRiskScore,
			&ra.Status, &ra.Notes,
			&ra.CreatedAt, &ra.UpdatedAt,
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
		SET assessment_date=$1, assessor_id=$2, overall_risk_score=$3,
		    risk_level=$4, security_risk_score=$5, financial_risk_score=$6,
		    operational_risk_score=$7, legal_risk_score=$8,
		    status=$9, notes=$10, updated_at=NOW()
		WHERE code=$11
	`, ra.AssessmentDate, ra.AssessorID,
		ra.OverallRiskScore, ra.RiskLevel,
		ra.SecurityRiskScore, ra.FinancialRiskScore,
		ra.OperationalRiskScore, ra.LegalRiskScore,
		ra.Status, ra.Notes, ra.Code)
	if err != nil {
		return err
	}
	if result.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}
	return nil
}

func (r *Repository) Delete(ctx context.Context, code string) error {
	result, err := r.pool.Exec(ctx, `DELETE FROM risk_assessments WHERE code = $1`, code)
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

	whereSQL := ""
	if len(where) > 0 {
		whereSQL = "WHERE " + strings.Join(where, " AND ")
	}

	join := "FROM risk_assessments ra JOIN vendors v ON ra.vendor_id = v.id"

	countSQL := fmt.Sprintf("SELECT COUNT(*) %s %s", join, whereSQL)
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
		SELECT ra.id, ra.code, ra.vendor_id, v.code,
			ra.assessment_date, ra.assessor_id,
			ra.overall_risk_score, ra.risk_level,
			ra.security_risk_score, ra.financial_risk_score,
			ra.operational_risk_score, ra.legal_risk_score,
			ra.status, ra.notes,
			ra.created_at, ra.updated_at
		%s
		%s
		ORDER BY ra.assessment_date DESC
		LIMIT $%d OFFSET $%d
	`, join, whereSQL, idx, idx+1)
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
			&ra.ID, &ra.Code, &ra.VendorID, &ra.VendorCode,
			&ra.AssessmentDate, &ra.AssessorID,
			&ra.OverallRiskScore, &ra.RiskLevel,
			&ra.SecurityRiskScore, &ra.FinancialRiskScore,
			&ra.OperationalRiskScore, &ra.LegalRiskScore,
			&ra.Status, &ra.Notes,
			&ra.CreatedAt, &ra.UpdatedAt,
		); err != nil {
			return nil, 0, err
		}
		assessments = append(assessments, ra)
	}
	return assessments, total, nil
}