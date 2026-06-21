package vendor_contact

import (
	"time"

	"github.com/google/uuid"
)

type VendorContact struct {
	ID        uuid.UUID `json:"id"`
	Code      string    `json:"code"`
	VendorID  uuid.UUID `json:"vendor_id"`
	Name      string    `json:"name"`
	Email     string    `json:"email,omitempty"`
	Phone     string    `json:"phone,omitempty"`
	CreatedAt time.Time `json:"created_at"`
}