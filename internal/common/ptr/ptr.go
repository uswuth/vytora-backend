// Package ptr provides utilities for converting between value and pointer types.
// This eliminates repetitive pointer conversion boilerplate across the codebase.
package ptr

import "github.com/google/uuid"

// Str returns a pointer to the given string value.
func Str(s string) *string { return &s }

// StrIfNotEmpty returns a pointer to the given string, or nil if it's empty.
func StrIfNotEmpty(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

// StrVal safely dereferences a *string, returning empty string if nil.
func StrVal(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

// UUIDStr returns a pointer to the string representation of a UUID, or nil if u is nil.
func UUIDStr(u *uuid.UUID) *string {
	if u == nil {
		return nil
	}
	s := u.String()
	return &s
}