package models

import (
	"time"

	"github.com/google/uuid"
)

type RiskAssessment struct {
	ID                  uuid.UUID `json:"id"`
	Code                string    `json:"code"`
	VendorID            uuid.UUID `json:"vendor_id"`
	VendorCode          string    `json:"vendor_code,omitempty"`
	AssessmentDate      time.Time `json:"assessment_date"`
	AssessorID          uuid.UUID `json:"assessor_id"`
	AssessorCode        string    `json:"assessor_code,omitempty"`
	OverallRiskScore    float64   `json:"overall_risk_score"`
	RiskLevel           string    `json:"risk_level"`
	SecurityRiskScore   float64   `json:"security_risk_score"`
	FinancialRiskScore  float64   `json:"financial_risk_score"`
	OperationalRiskScore float64  `json:"operational_risk_score"`
	LegalRiskScore      float64   `json:"legal_risk_score"`
	Status              string    `json:"status"`
	Notes               string    `json:"notes,omitempty"`
	CreatedAt           time.Time `json:"created_at"`
	UpdatedAt           time.Time `json:"updated_at"`
}