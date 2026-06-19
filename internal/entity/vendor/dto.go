package vendor

type CreateVendorRequest struct {
	Name          string `json:"name" validate:"required,min=2,max=255"`
	Category      string `json:"category" validate:"required,max=100"`
	ContactPerson string `json:"contact_person"`
	ContactEmail  string `json:"contact_email" validate:"omitempty,email"`
	Country       string `json:"country"`
	RiskLevel     string `json:"risk_level" validate:"required,oneof=Low Medium High Critical"`
}

type UpdateVendorRequest struct {
	Name          string `json:"name" validate:"required,min=2,max=255"`
	Category      string `json:"category" validate:"required,max=100"`
	ContactPerson string `json:"contact_person"`
	ContactEmail  string `json:"contact_email" validate:"omitempty,email"`
	Country       string `json:"country"`
	RiskLevel     string `json:"risk_level" validate:"required,oneof=Low Medium High Critical"`
	Status        string `json:"status" validate:"required,oneof=Draft Submitted RiskReview ComplianceReview Approved Rejected Active Inactive"`
}