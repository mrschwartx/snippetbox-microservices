package apperror

import (
	"errors"
	"maps"
	"slices"
	"strings"
)

var (
	ErrNotFound           = errors.New("record not found")
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrDuplicateEmail     = errors.New("duplicate email")
	ErrUnauthenticated    = errors.New("authentication required")
)

type ValidationError struct {
	FieldErrors    map[string]string
	NonFieldErrors []string
}

func NewFieldValidationError(field, message string) *ValidationError {
	return &ValidationError{FieldErrors: map[string]string{field: message}}
}

func (e *ValidationError) Error() string {
	fields := slices.Sorted(maps.Keys(e.FieldErrors))

	parts := make([]string, 0, len(fields)+len(e.NonFieldErrors))
	for _, field := range fields {
		parts = append(parts, field+": "+e.FieldErrors[field])
	}
	parts = append(parts, e.NonFieldErrors...)

	return "validation failed: " + strings.Join(parts, "; ")
}
