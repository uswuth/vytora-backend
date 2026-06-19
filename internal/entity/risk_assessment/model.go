package risk_assessment

import (
	"time"

	"github.com/google/uuid"
)

type RiskAssessment struct {
	ID                  uuid.UUID  `json:"id"`
	Code                string     `json:"code"`
	VendorID            uuid.UUID  `json:"vendor_id"`
	VendorCode          string     `json:"vendor_code,omitempty"`
	AssessmentDate      time.Time  `json:"assessment_date"`
	AssessorID          *uuid.UUID `json:"assessor_id,omitempty"`
	OverallRiskScore    float64    `json:"overall_risk_score"`
	RiskLevel           string     `json:"risk_level"`
	SecurityRiskScore   float64    `json:"security_risk_score"`
	FinancialRiskScore  float64    `json:"financial_risk_score"`
	OperationalRiskScore float64   `json:"operational_risk_score"`
	LegalRiskScore      float64    `json:"legal_risk_score"`
	Status              string     `json:"status"`
	Notes               string     `json:"notes,omitempty"`
	CreatedAt           time.Time  `json:"created_at"`
	UpdatedAt           time.Time  `json:"updated_at"`
}

type UpdateRiskAssessmentRequest struct {
	AssessmentDate       string  `json:"assessment_date" validate:"required"`
	OverallRiskScore     float64 `json:"overall_risk_score" validate:"required,min=0,max=100"`
	RiskLevel            string  `json:"risk_level" validate:"required,oneof=Low Medium High Critical"`
	SecurityRiskScore    float64 `json:"security_risk_score" validate:"required,min=0,max=100"`
	FinancialRiskScore   float64 `json:"financial_risk_score" validate:"required,min=0,max=100"`
	OperationalRiskScore float64 `json:"operational_risk_score" validate:"required,min=0,max=100"`
	LegalRiskScore       float64 `json:"legal_risk_score" validate:"required,min=0,max=100"`
	Status               string  `json:"status" validate:"required,oneof=Draft Reviewed Approved"`
	Notes                string  `json:"notes"`
}