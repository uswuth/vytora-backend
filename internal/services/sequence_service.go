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
	"category":          "CAT",
}

var DigitsMap = map[string]int{
	"user":              3, // USR001–USR999
	"vendor":            3, // VEN001–VEN999
	"vendor_contact":    4, // VCN0001–VCN9999
	"risk_assessment":   5, // RAK00001–RAK99999
	"compliance_record": 5, // CMP00001–CMP99999
	"contract":          5, // CTR00001–CTR99999
	"audit_trail":       6, // AUD000001–AUD999999
	"category":          3, // CAT001–CAT999
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
	digits, ok := DigitsMap[entityName]
	if !ok {
		digits = 3 // safe default
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
	
	code := fmt.Sprintf("%s%0*d", prefix, digits, nextVal)
	return code, nil
}
func formatCode(prefix string, digits int, val int) string {
	return fmt.Sprintf("%s%0*d", prefix, digits, val)
}