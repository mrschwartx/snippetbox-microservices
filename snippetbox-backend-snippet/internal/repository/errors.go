package repository

import (
	"errors"
	"fmt"
	"strings"

	"gorm.io/gorm"

	"snippetbox/backend-snippet/internal/apperror"
)

func isDuplicateKey(err error) bool {
	if errors.Is(err, gorm.ErrDuplicatedKey) {
		return true
	}

	message := strings.ToLower(err.Error())
	return strings.Contains(message, "duplicate entry") ||
		strings.Contains(message, "unique constraint failed") ||
		strings.Contains(message, "error 1062")
}

func findError(action string, err error) error {
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return apperror.ErrNotFound
	}
	return fmt.Errorf("repository: %s: %w", action, err)
}
