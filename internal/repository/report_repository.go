package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type ReportRepository struct {
	pool *pgxpool.Pool
}

func NewReportRepository(pool *pgxpool.Pool) *ReportRepository {
	return &ReportRepository{pool: pool}
}

type Summary struct {
	TotalVendors            int            `json:"total_vendors"`
	VendorsByStatus         map[string]int `json:"vendors_by_status"`
	VendorsByRiskLevel      map[string]int `json:"vendors_by_risk_level"`
	ExpiringContracts30     int            `json:"expiring_contracts_30_days"`
	ExpiringContracts60     int            `json:"expiring_contracts_60_days"`
	ExpiringContracts90     int            `json:"expiring_contracts_90_days"`
	ExpiringCompliance30    int            `json:"expiring_compliance_30_days"`
	ExpiringCompliance60    int            `json:"expiring_compliance_60_days"`
	ExpiringCompliance90    int            `json:"expiring_compliance_90_days"`
	PendingRiskAssessments  int            `json:"pending_risk_assessments"`
	ApprovedRiskAssessments int            `json:"approved_risk_assessments"`
}

type MonthlyOnboarding struct {
	Month string `json:"month"`
	Count int    `json:"count"`
}

// buildFilter returns (whereClause, args) and an index for the next placeholder.
func buildFilter(deptManagerID *uuid.UUID) (string, []interface{}, int) {
	if deptManagerID == nil {
		return "", nil, 1
	}
	return "WHERE assigned_dept_manager_id = $1", []interface{}{*deptManagerID}, 2
}

func (r *ReportRepository) GetSummary(ctx context.Context, deptManagerID *uuid.UUID) (*Summary, error) {
	s := &Summary{
		VendorsByStatus:    make(map[string]int),
		VendorsByRiskLevel: make(map[string]int),
	}

	whereClause, baseArgs, nextIdx := buildFilter(deptManagerID)

	// Total vendors
	query := fmt.Sprintf(`SELECT COUNT(*) FROM vendors %s`, whereClause)
	err := r.pool.QueryRow(ctx, query, baseArgs...).Scan(&s.TotalVendors)
	if err != nil {
		return nil, err
	}

	// By status
	query = fmt.Sprintf(`SELECT status, COUNT(*) FROM vendors %s GROUP BY status`, whereClause)
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

	// By risk level
	query = fmt.Sprintf(`SELECT risk_level, COUNT(*) FROM vendors %s GROUP BY risk_level`, whereClause)
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

	// Expiring contracts (30/60/90 days)
	for _, days := range []int{30, 60, 90} {
		cutoff := time.Now().AddDate(0, 0, days)
		contractWhere := fmt.Sprintf(`c.end_date <= $%d AND c.end_date >= CURRENT_DATE`, nextIdx)
		query = fmt.Sprintf(`SELECT COUNT(*) FROM contracts c JOIN vendors v ON c.vendor_id = v.id %s`, whereClause)
		if whereClause != "" {
			query += " AND " + contractWhere
		} else {
			query += "WHERE " + contractWhere
		}
		args := append(baseArgs, cutoff)
		err = r.pool.QueryRow(ctx, query, args...).Scan(getExpiringContractField(days, s))
		if err != nil {
			return nil, err
		}
	}

	// Expiring compliance (30/60/90 days)
	for _, days := range []int{30, 60, 90} {
		cutoff := time.Now().AddDate(0, 0, days)
		compWhere := fmt.Sprintf(`cr.status = 'Approved' AND cr.valid_until <= $%d AND cr.valid_until >= CURRENT_DATE`, nextIdx)
		query = fmt.Sprintf(`SELECT COUNT(*) FROM compliance_records cr JOIN vendors v ON cr.vendor_id = v.id %s`, whereClause)
		if whereClause != "" {
			query += " AND " + compWhere
		} else {
			query += "WHERE " + compWhere
		}
		args := append(baseArgs, cutoff)
		err = r.pool.QueryRow(ctx, query, args...).Scan(getExpiringComplianceField(days, s))
		if err != nil {
			return nil, err
		}
	}

	// Risk assessment counts
	query = fmt.Sprintf(`SELECT COUNT(*) FROM risk_assessments ra JOIN vendors v ON ra.vendor_id = v.id %s`, whereClause)
	if whereClause != "" {
		query += " AND ra.status = 'Draft'"
	} else {
		query += "WHERE ra.status = 'Draft'"
	}
	err = r.pool.QueryRow(ctx, query, baseArgs...).Scan(&s.PendingRiskAssessments)
	if err != nil {
		return nil, err
	}

	query = fmt.Sprintf(`SELECT COUNT(*) FROM risk_assessments ra JOIN vendors v ON ra.vendor_id = v.id %s`, whereClause)
	if whereClause != "" {
		query += " AND ra.status = 'Approved'"
	} else {
		query += "WHERE ra.status = 'Approved'"
	}
	err = r.pool.QueryRow(ctx, query, baseArgs...).Scan(&s.ApprovedRiskAssessments)
	if err != nil {
		return nil, err
	}

	return s, nil
}

func getExpiringContractField(days int, s *Summary) *int {
	switch days {
	case 30:
		return &s.ExpiringContracts30
	case 60:
		return &s.ExpiringContracts60
	default:
		return &s.ExpiringContracts90
	}
}

func getExpiringComplianceField(days int, s *Summary) *int {
	switch days {
	case 30:
		return &s.ExpiringCompliance30
	case 60:
		return &s.ExpiringCompliance60
	default:
		return &s.ExpiringCompliance90
	}
}

func (r *ReportRepository) GetMonthlyOnboarding(ctx context.Context, deptManagerID *uuid.UUID) ([]MonthlyOnboarding, error) {
	whereClause, args, _ := buildFilter(deptManagerID)
	query := fmt.Sprintf(`SELECT TO_CHAR(created_at, 'YYYY-MM') AS month, COUNT(*) FROM vendors %s GROUP BY month ORDER BY month`, whereClause)
	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []MonthlyOnboarding
	for rows.Next() {
		var m MonthlyOnboarding
		if err := rows.Scan(&m.Month, &m.Count); err != nil {
			return nil, err
		}
		result = append(result, m)
	}
	return result, nil
}