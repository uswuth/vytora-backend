package services

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

// PrefixMap maps entity names to their code prefix.
var PrefixMap = map[string]string{
	"user":              "USR",
	"vendor":            "VEN",
	"vendor_contact":    "VCN",
	"risk_assessment":   "RAK",
	"compliance_record": "CMP",
	"contract":          "CTR",
	"audit_trail":       "AUD",
}

// SequenceService generates entity codes like VEN001.
type SequenceService struct {
	pool *pgxpool.Pool
}

// NewSequenceService creates a new SequenceService.
func NewSequenceService(pool *pgxpool.Pool) *SequenceService {
	return &SequenceService{pool: pool}
}

// NextCode generates the next code for the given entity name.
// Format: PREFIX + 3-digit number (e.g., VEN001, USR042)
func (s *SequenceService) NextCode(ctx context.Context, entityName string) (string, error) {
	prefix, ok := PrefixMap[entityName]
	if !ok {
		return "", fmt.Errorf("unknown entity name: %s", entityName)
	}

	var nextVal int
	err := s.pool.QueryRow(ctx, `
		UPDATE entity_sequences
		SET next_value = next_value + 1
		WHERE entity_name = $1
		RETURNING next_value - 1;
	`, entityName).Scan(&nextVal)
	if err != nil {
		return "", fmt.Errorf("failed to fetch next sequence: %w", err)
	}

	code := fmt.Sprintf("%s%03d", prefix, nextVal)
	return code, nil
}
