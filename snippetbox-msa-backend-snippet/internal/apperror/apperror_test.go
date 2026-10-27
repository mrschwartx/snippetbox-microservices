package apperror_test

import (
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"

	"snippetbox/backend-snippet/internal/apperror"
)

func TestValidationError_Error(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		err  *apperror.ValidationError
		want string
	}{
		{
			name: "single field",
			err:  apperror.NewFieldValidationError("email", "taken"),
			want: "validation failed: email: taken",
		},
		{
			name: "fields are sorted for a stable message",
			err: &apperror.ValidationError{FieldErrors: map[string]string{
				"title": "blank", "content": "blank", "expires": "bad",
			}},
			want: "validation failed: content: blank; expires: bad; title: blank",
		},
		{
			name: "non field errors follow field errors",
			err: &apperror.ValidationError{
				FieldErrors:    map[string]string{"name": "blank"},
				NonFieldErrors: []string{"try again"},
			},
			want: "validation failed: name: blank; try again",
		},
		{
			name: "empty",
			err:  &apperror.ValidationError{},
			want: "validation failed: ",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.want, tt.err.Error())
		})
	}
}

func TestSentinelErrors_SurviveWrapping(t *testing.T) {
	t.Parallel()

	sentinels := []error{
		apperror.ErrNotFound,
		apperror.ErrUnauthenticated,
	}

	for _, sentinel := range sentinels {
		wrapped := fmt.Errorf("layer: %w", sentinel)
		assert.ErrorIs(t, wrapped, sentinel)
		for _, other := range sentinels {
			if other != sentinel {
				assert.False(t, errors.Is(wrapped, other))
			}
		}
	}
}

func TestValidationError_ExtractableWithErrorsAs(t *testing.T) {
	t.Parallel()

	err := fmt.Errorf("service: %w", apperror.NewFieldValidationError("name", "blank"))

	var validation *apperror.ValidationError
	assert.ErrorAs(t, err, &validation)
	assert.Equal(t, "blank", validation.FieldErrors["name"])
}
