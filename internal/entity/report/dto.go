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

type HighRiskVendorItem struct {
	Code                string  `json:"code"`
	Name                string  `json:"name"`
	Category            string  `json:"category"`
	RiskLevel           string  `json:"risk_level"`
	OverallRiskScore    float64 `json:"overall_risk_score,omitempty"`
	LatestAssessment    string  `json:"latest_assessment_date,omitempty"`
	AssessmentStatus    string  `json:"assessment_status,omitempty"`
	ExpiringContracts30 int     `json:"expiring_contracts_30_days,omitempty"`
}

type ExpiringContractItem struct {
	Code           string  `json:"code"`
	VendorName     string  `json:"vendor_name"`
	VendorCode     string  `json:"vendor_code"`
	ContractNumber string  `json:"contract_number"`
	EndDate        string  `json:"end_date"`
	DaysRemaining  int     `json:"days_remaining"`
	RenewalStatus  string  `json:"renewal_status"`
	ContractValue  float64 `json:"contract_value,omitempty"`
}

type ComplianceSummaryItem struct {
	CertificationType string `json:"certification_type"`
	Approved          int    `json:"approved"`
	Pending           int    `json:"pending"`
	Expired           int    `json:"expired"`
	Total             int    `json:"total"`
}

type TimeSeriesPoint struct {
	Date  string `json:"date"`
	Count int    `json:"count"`
}

type TimeSeriesResponse struct {
	Vendors         []TimeSeriesPoint `json:"vendors"`
	RiskAssessments []TimeSeriesPoint `json:"risk_assessments"`
	Compliance      []TimeSeriesPoint `json:"compliance_records"`
}

// ExportRowVendor is a flat CSV-friendly vendor row.
type ExportRowVendor struct {
	Code      string `json:"code"`
	Name      string `json:"name"`
	Category  string `json:"category"`
	RiskLevel string `json:"risk_level"`
	Status    string `json:"status"`
	Country   string `json:"country"`
	CreatedBy string `json:"created_by"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

// ExportRowRisk is a flat CSV-friendly risk assessment row.
type ExportRowRisk struct {
	Code             string  `json:"code"`
	VendorCode       string  `json:"vendor_code"`
	AssessmentDate   string  `json:"assessment_date"`
	OverallRiskScore float64 `json:"overall_risk_score"`
	RiskLevel        string  `json:"risk_level"`
	SecurityScore    float64 `json:"security_risk_score"`
	FinancialScore   float64 `json:"financial_risk_score"`
	OperationalScore float64 `json:"operational_risk_score"`
	LegalScore       float64 `json:"legal_risk_score"`
	Status           string  `json:"status"`
}

// ExportRowCompliance is a flat CSV-friendly compliance row.
type ExportRowCompliance struct {
	Code              string `json:"code"`
	VendorCode        string `json:"vendor_code"`
	CertificationType string `json:"certification_type"`
	Status            string `json:"status"`
	ValidFrom         string `json:"valid_from"`
	ValidUntil        string `json:"valid_until"`
	IssuedBy          string `json:"issued_by"`
}

// ExportRowContract is a flat CSV-friendly contract row.
type ExportRowContract struct {
	Code           string  `json:"code"`
	VendorCode     string  `json:"vendor_code"`
	ContractNumber string  `json:"contract_number"`
	StartDate      string  `json:"start_date"`
	EndDate        string  `json:"end_date"`
	ContractValue  float64 `json:"contract_value,omitempty"`
	RenewalStatus  string  `json:"renewal_status"`
}