package models

import (
	"time"

	"github.com/google/uuid"
)

type AuditTrail struct {
	ID        uuid.UUID `json:"id"`
	Code      string    `json:"code"`
	TableName string    `json:"table_name"`
	RecordID  uuid.UUID `json:"record_id"`
	Action    string    `json:"action"`
	FieldName string    `json:"field_name,omitempty"`
	OldValue  string    `json:"old_value,omitempty"`
	NewValue  string    `json:"new_value,omitempty"`
	ChangedBy uuid.UUID `json:"changed_by"`
	ChangedAt time.Time `json:"changed_at"`
}
