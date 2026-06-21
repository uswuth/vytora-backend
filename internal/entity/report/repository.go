package report

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository struct {
	pool *pgxpool.Pool
}

func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{pool: pool}
}

func buildFilter(deptManagerID *uuid.UUID) (string, []interface{}, int) {
	if deptManagerID == nil {
		return "", nil, 1
	}
	return "WHERE v.assigned_dept_manager_id = $1", []interface{}{*deptManagerID}, 2
}

func buildCreatedByFilter(createdBy *uuid.UUID) (string, []interface{}, int) {
	if createdBy == nil {
		return "", nil, 1
	}
	return "v.created_by = $1", []interface{}{*createdBy}, 2
}

func (r *Repository) GetSummary(ctx context.Context, deptManagerID *uuid.UUID) (*SummaryResponse, error) {
	s := &SummaryResponse{
		VendorsByStatus:    make(map[string]int),
		VendorsByRiskLevel: make(map[string]int),
	}

	whereClause, baseArgs, nextIdx := buildFilter(deptManagerID)

	// Total vendors
	query := fmt.Sprintf(`SELECT COUNT(*) FROM vendors v %s`, whereClause)
	err := r.pool.QueryRow(ctx, query, baseArgs...).Scan(&s.TotalVendors)
	if err != nil {
		return nil, err
	}

	// By status
	query = fmt.Sprintf(`SELECT v.status, COUNT(*) FROM vendors v %s GROUP BY v.status`, whereClause)
	rows, err := r.pool.Query(ctx, query, baseArgs...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var status string
		var count int
		if err := rows.Scan(&status, &count); err != nil {
			return nil, err
		}
		s.VendorsByStatus[status] = count
	}
	rows.Close()

	// By risk level
	query = fmt.Sprintf(`SELECT v.risk_level, COUNT(*) FROM vendors v %s GROUP BY v.risk_level`, whereClause)
	rows, err = r.pool.Query(ctx, query, baseArgs...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var level string
		var count int
		if err := rows.Scan(&level, &count); err != nil {
			return nil, err
		}
		s.VendorsByRiskLevel[level] = count
	}
	rows.Close()

	// Expiring contracts (30/60/90 days)
	for _, days := range []int{30, 60, 90} {
		cutoff := time.Now().AddDate(0, 0, days)
		var cnt int
		contractWhere := fmt.Sprintf(`c.end_date <= $%d AND c.end_date >= CURRENT_DATE`, nextIdx)
		q := fmt.Sprintf(`SELECT COUNT(*) FROM contracts c JOIN vendors v ON c.vendor_id = v.id %s`, whereClause)
		if whereClause != "" {
			q += " AND " + contractWhere
		} else {
			q += "WHERE " + contractWhere
		}
		args := append(baseArgs, cutoff)
		err = r.pool.QueryRow(ctx, q, args...).Scan(&cnt)
		if err != nil {
			return nil, err
		}
		switch days {
		case 30:
			s.ExpiringContracts30 = cnt
		case 60:
			s.ExpiringContracts60 = cnt
		default:
			s.ExpiringContracts90 = cnt
		}
	}

	// Expiring compliance (30/60/90 days)
	for _, days := range []int{30, 60, 90} {
		cutoff := time.Now().AddDate(0, 0, days)
		var cnt int
		compWhere := fmt.Sprintf(`cr.status = 'Approved' AND cr.valid_until <= $%d AND cr.valid_until >= CURRENT_DATE`, nextIdx)
		q := fmt.Sprintf(`SELECT COUNT(*) FROM compliance_records cr JOIN vendors v ON cr.vendor_id = v.id %s`, whereClause)
		if whereClause != "" {
			q += " AND " + compWhere
		} else {
			q += "WHERE " + compWhere
		}
		args := append(baseArgs, cutoff)
		err = r.pool.QueryRow(ctx, q, args...).Scan(&cnt)
		if err != nil {
			return nil, err
		}
		switch days {
		case 30:
			s.ExpiringCompliance30 = cnt
		case 60:
			s.ExpiringCompliance60 = cnt
		default:
			s.ExpiringCompliance90 = cnt
		}
	}

	// Risk assessment counts
	q := fmt.Sprintf(`SELECT COUNT(*) FROM risk_assessments ra JOIN vendors v ON ra.vendor_id = v.id %s`, whereClause)
	if whereClause != "" {
		q += " AND ra.status = 'Draft'"
	} else {
		q += "WHERE ra.status = 'Draft'"
	}
	err = r.pool.QueryRow(ctx, q, baseArgs...).Scan(&s.PendingRiskAssessments)
	if err != nil {
		return nil, err
	}

	q = fmt.Sprintf(`SELECT COUNT(*) FROM risk_assessments ra JOIN vendors v ON ra.vendor_id = v.id %s`, whereClause)
	if whereClause != "" {
		q += " AND ra.status = 'Approved'"
	} else {
		q += "WHERE ra.status = 'Approved'"
	}
	err = r.pool.QueryRow(ctx, q, baseArgs...).Scan(&s.ApprovedRiskAssessments)
	if err != nil {
		return nil, err
	}

	return s, nil
}

func (r *Repository) GetMonthlyOnboarding(ctx context.Context, deptManagerID *uuid.UUID) ([]MonthlyOnboardingItem, error) {
	whereClause, args, _ := buildFilter(deptManagerID)
	query := fmt.Sprintf(`SELECT TO_CHAR(v.created_at, 'YYYY-MM') AS month, COUNT(*) FROM vendors v %s GROUP BY month ORDER BY month`, whereClause)
	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []MonthlyOnboardingItem
	for rows.Next() {
		var m MonthlyOnboardingItem
		if err := rows.Scan(&m.Month, &m.Count); err != nil {
			return nil, err
		}
		result = append(result, m)
	}
	return result, nil
}

// GetHighRiskVendors returns vendors with High/Critical risk level and their latest assessment.
func (r *Repository) GetHighRiskVendors(ctx context.Context, createdBy *uuid.UUID) ([]HighRiskVendorItem, error) {
	filterClause, args, _ := buildCreatedByFilter(createdBy)

	whereClause := "WHERE v.risk_level IN ('High', 'Critical')"
	if filterClause != "" {
		whereClause = "WHERE " + filterClause + " AND v.risk_level IN ('High', 'Critical')"
	}

	query := fmt.Sprintf(`
		SELECT v.code, v.name, v.category, v.risk_level,
			COALESCE(ra.overall_risk_score, 0) AS overall_risk_score,
			COALESCE(ra.assessment_date::TEXT, '') AS latest_assessment,
			COALESCE(ra.status, '') AS assessment_status,
			(SELECT COUNT(*) FROM contracts c WHERE c.vendor_id = v.id AND c.end_date BETWEEN CURRENT_DATE AND CURRENT_DATE + INTERVAL '30 days') AS expiring_30
		FROM vendors v
		LEFT JOIN LATERAL (
			SELECT overall_risk_score, assessment_date, status
			FROM risk_assessments
			WHERE vendor_id = v.id
			ORDER BY assessment_date DESC
			LIMIT 1
		) ra ON TRUE
		%s
		ORDER BY v.risk_level DESC, v.name ASC
	`, whereClause)

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []HighRiskVendorItem
	for rows.Next() {
		var item HighRiskVendorItem
		if err := rows.Scan(
			&item.Code, &item.Name, &item.Category, &item.RiskLevel,
			&item.OverallRiskScore, &item.LatestAssessment, &item.AssessmentStatus,
			&item.ExpiringContracts30,
		); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, nil
}

// GetExpiringContractsReport returns contracts expiring soon with days remaining.
func (r *Repository) GetExpiringContractsReport(ctx context.Context, createdBy *uuid.UUID) ([]ExpiringContractItem, error) {
	filterClause, args, _ := buildCreatedByFilter(createdBy)

	whereClause := "WHERE c.end_date >= CURRENT_DATE"
	join := "JOIN vendors v ON c.vendor_id = v.id"
	if filterClause != "" {
		whereClause = "WHERE " + filterClause + " AND c.end_date >= CURRENT_DATE"
	}

	query := fmt.Sprintf(`
		SELECT c.code, v.name, v.code, c.contract_number,
			c.end_date::TEXT,
			(c.end_date - CURRENT_DATE) AS days_remaining,
			c.renewal_status,
			COALESCE(c.contract_value, 0)
		FROM contracts c
		%s
		%s
		ORDER BY c.end_date ASC
	`, join, whereClause)

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []ExpiringContractItem
	for rows.Next() {
		var item ExpiringContractItem
		if err := rows.Scan(
			&item.Code, &item.VendorName, &item.VendorCode, &item.ContractNumber,
			&item.EndDate, &item.DaysRemaining, &item.RenewalStatus, &item.ContractValue,
		); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, nil
}

// GetComplianceSummary returns compliance status counts grouped by certification type.
func (r *Repository) GetComplianceSummary(ctx context.Context, createdBy *uuid.UUID) ([]ComplianceSummaryItem, error) {
	filterClause, args, _ := buildCreatedByFilter(createdBy)

	var where string
	if filterClause != "" {
		where = "WHERE " + filterClause
	}

	query := fmt.Sprintf(`
		SELECT cr.certification_type,
			COUNT(*) FILTER (WHERE cr.status = 'Approved') AS approved,
			COUNT(*) FILTER (WHERE cr.status = 'Pending') AS pending,
			COUNT(*) FILTER (WHERE cr.status = 'Expired') AS expired,
			COUNT(*) AS total
		FROM compliance_records cr
		JOIN vendors v ON cr.vendor_id = v.id
		%s
		GROUP BY cr.certification_type
		ORDER BY cr.certification_type
	`, where)

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []ComplianceSummaryItem
	for rows.Next() {
		var item ComplianceSummaryItem
		if err := rows.Scan(
			&item.CertificationType, &item.Approved, &item.Pending, &item.Expired, &item.Total,
		); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, nil
}

// GetTimeSeries returns monthly counts for vendors, risk assessments, and compliance records.
func (r *Repository) GetTimeSeries(ctx context.Context, createdBy *uuid.UUID) (*TimeSeriesResponse, error) {
	resp := &TimeSeriesResponse{}

	whereClause, args, _ := buildCreatedByFilter(createdBy)
	whereSQL := ""
	if whereClause != "" {
		whereSQL = "WHERE " + whereClause
	}

	// Vendor time series
	vendorQuery := fmt.Sprintf(`SELECT TO_CHAR(created_at, 'YYYY-MM') AS month, COUNT(*) FROM vendors v %s GROUP BY month ORDER BY month`, whereSQL)
	rows, err := r.pool.Query(ctx, vendorQuery, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var p TimeSeriesPoint
		if err := rows.Scan(&p.Date, &p.Count); err != nil {
			return nil, err
		}
		resp.Vendors = append(resp.Vendors, p)
	}
	rows.Close()

	// Risk assessment time series
	raQuery := `SELECT TO_CHAR(created_at, 'YYYY-MM') AS month, COUNT(*) FROM risk_assessments GROUP BY month ORDER BY month`
	rows, err = r.pool.Query(ctx, raQuery)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var p TimeSeriesPoint
		if err := rows.Scan(&p.Date, &p.Count); err != nil {
			return nil, err
		}
		resp.RiskAssessments = append(resp.RiskAssessments, p)
	}
	rows.Close()

	// Compliance time series
	compQuery := `SELECT TO_CHAR(created_at, 'YYYY-MM') AS month, COUNT(*) FROM compliance_records GROUP BY month ORDER BY month`
	rows, err = r.pool.Query(ctx, compQuery)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var p TimeSeriesPoint
		if err := rows.Scan(&p.Date, &p.Count); err != nil {
			return nil, err
		}
		resp.Compliance = append(resp.Compliance, p)
	}

	return resp, nil
}