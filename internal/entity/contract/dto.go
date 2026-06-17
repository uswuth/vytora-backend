package contract

type CreateContractRequest struct {
	VendorCode     string   `json:"vendor_code" validate:"required"`
	ContractNumber string   `json:"contract_number" validate:"required,max=100"`
	StartDate      string   `json:"start_date" validate:"required"` // YYYY-MM-DD
	EndDate        string   `json:"end_date" validate:"required"`
	ContractValue  *float64 `json:"contract_value"`
	RenewalStatus  string   `json:"renewal_status" validate:"required,oneof=Auto-Renew Manual Expiring"`
}