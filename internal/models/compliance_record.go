package models

import (
	"time"

	"github.com/google/uuid"
)

type ComplianceRecord struct {
	ID                uuid.UUID  `json:"id"`
	Code              string     `json:"code"`
	VendorID          uuid.UUID  `json:"vendor_id"`
	VendorCode        string     `json:"vendor_code,omitempty"`
	CertificationType string     `json:"certification_type"`
	Status            string     `json:"status"`
	ValidFrom         *time.Time `json:"valid_from,omitempty"`
	ValidUntil        *time.Time `json:"valid_until,omitempty"`
	IssuedBy          string     `json:"issued_by,omitempty"`
	EvidenceURL       string     `json:"evidence_url,omitempty"`
	ReviewedBy        *uuid.UUID `json:"reviewed_by,omitempty"`
	CreatedAt         time.Time  `json:"created_at"`
	UpdatedAt         time.Time  `json:"updated_at"`
}