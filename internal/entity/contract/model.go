package contract

import (
	"time"

	"github.com/google/uuid"
)

type Contract struct {
	ID             uuid.UUID  `json:"id"`
	Code           string     `json:"code"`
	VendorID       uuid.UUID  `json:"vendor_id"`
	VendorCode     string     `json:"vendor_code,omitempty"`
	ContractNumber string     `json:"contract_number"`
	StartDate      time.Time  `json:"start_date"`
	EndDate        time.Time  `json:"end_date"`
	ContractValue  *float64   `json:"contract_value,omitempty"`
	RenewalStatus  string     `json:"renewal_status"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
}