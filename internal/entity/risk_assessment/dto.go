package risk_assessment

type CreateRiskAssessmentRequest struct {
	VendorCode           string  `json:"vendor_code" validate:"required"`
	AssessmentDate       string  `json:"assessment_date" validate:"required"`
	OverallRiskScore     float64 `json:"overall_risk_score" validate:"required,min=0,max=100"`
	RiskLevel            string  `json:"risk_level" validate:"required,oneof=Low Medium High Critical"`
	SecurityRiskScore    float64 `json:"security_risk_score" validate:"required,min=0,max=100"`
	FinancialRiskScore   float64 `json:"financial_risk_score" validate:"required,min=0,max=100"`
	OperationalRiskScore float64 `json:"operational_risk_score" validate:"required,min=0,max=100"`
	LegalRiskScore       float64 `json:"legal_risk_score" validate:"required,min=0,max=100"`
	Notes                string  `json:"notes"`
}

type ListResponse struct {
	Data  []RiskAssessment `json:"data"`
	Total int              `json:"total"`
}