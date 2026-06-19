package vendor

import (
	"time"

	"github.com/google/uuid"
)

type Vendor struct {
	ID                    uuid.UUID  `json:"id"`
	Code                  string     `json:"code"`
	Name                  string     `json:"name"`
	Category              string     `json:"category"`
	ContactPerson         string     `json:"contact_person,omitempty"`
	ContactEmail          string     `json:"contact_email,omitempty"`
	Country               string     `json:"country,omitempty"`
	ContractStartDate     *time.Time `json:"contract_start_date,omitempty"`
	ContractEndDate       *time.Time `json:"contract_end_date,omitempty"`
	RiskLevel             string     `json:"risk_level"`
	Status                string     `json:"status"`
	AssignedDeptManagerID *uuid.UUID `json:"assigned_dept_manager_id,omitempty"`
	CreatedBy             uuid.UUID  `json:"created_by"`
	CreatedAt             time.Time  `json:"created_at"`
	UpdatedAt             time.Time  `json:"updated_at"`
}
