package risk_assessment

import (
	"time"

	"github.com/google/uuid"
)

type RiskAssessment struct {
	ID              uuid.UUID `json:"id"`
	Code            string    `json:"code"`
	VendorID        uuid.UUID `json:"vendor_id"`
	VendorCode      string    `json:"vendor_code"`
	AssessmentDate  time.Time `json:"assessment_date"`
	RiskLevel       string    `json:"risk_level"`
	Findings        string    `json:"findings"`
	Recommendations string    `json:"recommendations"`
	Status          string    `json:"status"`
	ReviewedBy      *uuid.UUID    `json:"reviewed_by,omitempty"`
	ReviewedAt      *time.Time     `json:"reviewed_at,omitempty"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

type UpdateRiskAssessmentRequest struct {
	RiskLevel       string `json:"risk_level" validate:"required,oneof=Low Medium High Critical"`
	Findings        string `json:"findings" validate:"required"`
	Recommendations string `json:"recommendations"`
	Status          string `json:"status" validate:"required,oneof=Pending Approved Rejected"`
}