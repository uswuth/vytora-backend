package audit_trail

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
)

// NextCodeFunc generates the next entity code.
type NextCodeFunc func(ctx context.Context, entity string) (string, error)

// Logger provides automatic audit trail logging for entity mutations.
type Logger struct {
	repo     *Repository
	nextCode NextCodeFunc
}

func NewLogger(repo *Repository, nextCode NextCodeFunc) *Logger {
	return &Logger{repo: repo, nextCode: nextCode}
}

// LogCreate logs a CREATE action with the full record as JSON new_value.
func (l *Logger) LogCreate(ctx context.Context, tableName string, recordID uuid.UUID, changedBy uuid.UUID, record interface{}) error {
	code, err := l.nextCode(ctx, "audit_trail")
	if err != nil {
		return fmt.Errorf("generate audit code: %w", err)
	}
	val, _ := json.Marshal(record)
	return l.repo.Create(ctx, &AuditTrail{
		Code:      code,
		TableName: tableName,
		RecordID:  recordID,
		Action:    "CREATE",
		FieldName: "",
		OldValue:  "",
		NewValue:  string(val),
		ChangedBy: changedBy,
	})
}

// LogCreateSimple logs a CREATE action with just the code as new_value.
func (l *Logger) LogCreateSimple(ctx context.Context, tableName string, recordID uuid.UUID, recordCode string, changedBy uuid.UUID) error {
	code, err := l.nextCode(ctx, "audit_trail")
	if err != nil {
		return fmt.Errorf("generate audit code: %w", err)
	}
	return l.repo.Create(ctx, &AuditTrail{
		Code:      code,
		TableName: tableName,
		RecordID:  recordID,
		Action:    "CREATE",
		FieldName: "",
		OldValue:  "",
		NewValue:  recordCode,
		ChangedBy: changedBy,
	})
}

// LogUpdateField logs a single field change during an UPDATE.
func (l *Logger) LogUpdateField(ctx context.Context, tableName string, recordID uuid.UUID, fieldName, oldValue, newValue string, changedBy uuid.UUID) error {
	code, err := l.nextCode(ctx, "audit_trail")
	if err != nil {
		return fmt.Errorf("generate audit code: %w", err)
	}
	return l.repo.Create(ctx, &AuditTrail{
		Code:      code,
		TableName: tableName,
		RecordID:  recordID,
		Action:    "UPDATE",
		FieldName: fieldName,
		OldValue:  oldValue,
		NewValue:  newValue,
		ChangedBy: changedBy,
	})
}

// LogDelete logs a DELETE action with the full record as JSON old_value.
func (l *Logger) LogDelete(ctx context.Context, tableName string, recordID uuid.UUID, record interface{}, changedBy uuid.UUID) error {
	code, err := l.nextCode(ctx, "audit_trail")
	if err != nil {
		return fmt.Errorf("generate audit code: %w", err)
	}
	val, _ := json.Marshal(record)
	return l.repo.Create(ctx, &AuditTrail{
		Code:      code,
		TableName: tableName,
		RecordID:  recordID,
		Action:    "DELETE",
		FieldName: "",
		OldValue:  string(val),
		NewValue:  "",
		ChangedBy: changedBy,
	})
}

// LogDeleteSimple logs a DELETE action with just the code as old_value.
func (l *Logger) LogDeleteSimple(ctx context.Context, tableName string, recordID uuid.UUID, recordCode string, changedBy uuid.UUID) error {
	code, err := l.nextCode(ctx, "audit_trail")
	if err != nil {
		return fmt.Errorf("generate audit code: %w", err)
	}
	return l.repo.Create(ctx, &AuditTrail{
		Code:      code,
		TableName: tableName,
		RecordID:  recordID,
		Action:    "DELETE",
		FieldName: "",
		OldValue:  recordCode,
		NewValue:  "",
		ChangedBy: changedBy,
	})
}