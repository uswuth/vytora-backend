package report

type SummaryResponse struct {
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

type MonthlyOnboardingItem struct {
	Month string `json:"month"`
	Count int    `json:"count"`
}