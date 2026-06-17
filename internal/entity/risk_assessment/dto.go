package risk_assessment

type CreateRiskAssessmentRequest struct {
	VendorCode      string `json:"vendor_code" validate:"required"`
	AssessmentDate  string `json:"assessment_date" validate:"required"`
	RiskLevel       string `json:"risk_level" validate:"required,oneof=Low Medium High Critical"`
	Findings        string `json:"findings" validate:"required"`
	Recommendations string `json:"recommendations"`
}

type ListResponse struct {
	Data  []RiskAssessment `json:"data"`
	Total int              `json:"total"`
}