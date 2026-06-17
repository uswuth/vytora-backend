package compliance_record

type CreateComplianceRequest struct {
	VendorCode        string `json:"vendor_code" validate:"required"`
	CertificationType string `json:"certification_type" validate:"required,oneof=ISO27001 SOC2 GDPR PCI_DSS"`
	ValidFrom         string `json:"valid_from" validate:"required"`      // YYYY-MM-DD
	ValidUntil        string `json:"valid_until" validate:"required"`
	IssuedBy          string `json:"issued_by"`
	EvidenceURL       string `json:"evidence_url"`
}

type UpdateComplianceRequest struct {
	CertificationType string `json:"certification_type" validate:"required,oneof=ISO27001 SOC2 GDPR PCI_DSS"`
	Status            string `json:"status" validate:"required,oneof=Pending Approved Expired"`
	ValidFrom         string `json:"valid_from"`
	ValidUntil        string `json:"valid_until"`
	IssuedBy          string `json:"issued_by"`
	EvidenceURL       string `json:"evidence_url"`
}