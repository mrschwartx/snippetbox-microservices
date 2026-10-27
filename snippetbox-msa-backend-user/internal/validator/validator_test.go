package validator_test

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"snippetbox/backend-user/internal/apperror"
	"snippetbox/backend-user/internal/validator"
)

func TestNotBlank(t *testing.T) {
	t.Parallel()

	assert.True(t, validator.NotBlank("a"))
	assert.False(t, validator.NotBlank(""))
	assert.False(t, validator.NotBlank(" \t\n"))
}

func TestLengthRules(t *testing.T) {
	t.Parallel()

	assert.True(t, validator.MaxChars("héllo", 5))
	assert.False(t, validator.MaxChars("héllo!", 5))
	assert.True(t, validator.MinChars("héllo", 5))
	assert.False(t, validator.MinChars("héll", 5))
	assert.True(t, validator.MaxBytes("abc", 3))
	assert.False(t, validator.MaxBytes("é", 1))
}

func TestMatches(t *testing.T) {
	t.Parallel()

	tests := []struct {
		email string
		want  bool
	}{
		{"user@example.com", true},
		{"first.last+tag@sub.example.co", true},
		{"user@localhost", true},
		{"", false},
		{"no-at-sign", false},
		{"@example.com", false},
		{"user@", false},
		{"user@-bad.com", false},
		{"us er@example.com", false},
	}

	for _, tt := range tests {
		t.Run(tt.email, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.want, validator.Matches(tt.email, validator.EmailRX))
		})
	}
}

func TestPermittedValue(t *testing.T) {
	t.Parallel()

	assert.True(t, validator.PermittedValue(7, 1, 7, 365))
	assert.False(t, validator.PermittedValue(2, 1, 7, 365))
	assert.True(t, validator.PermittedValue("a", "a"))
	assert.False(t, validator.PermittedValue("a"))
}

func TestValidator_Err(t *testing.T) {
	t.Parallel()

	t.Run("valid returns untyped nil", func(t *testing.T) {
		t.Parallel()
		var v validator.Validator
		v.CheckField(true, "name", "bad")
		assert.True(t, v.Valid())
		assert.NoError(t, v.Err())
	})

	t.Run("first message per field wins", func(t *testing.T) {
		t.Parallel()
		var v validator.Validator
		v.CheckField(false, "name", "first")
		v.CheckField(false, "name", "second")
		v.AddNonFieldError("global")

		assert.False(t, v.Valid())

		var verr *apperror.ValidationError
		require.ErrorAs(t, v.Err(), &verr)
		assert.Equal(t, map[string]string{"name": "first"}, verr.FieldErrors)
		assert.Equal(t, []string{"global"}, verr.NonFieldErrors)
	})

	t.Run("only non field errors", func(t *testing.T) {
		t.Parallel()
		var v validator.Validator
		v.AddNonFieldError("global")
		assert.False(t, v.Valid())
		assert.Error(t, v.Err())
	})

	t.Run("returned error is a snapshot", func(t *testing.T) {
		t.Parallel()
		var v validator.Validator
		v.AddFieldError("a", "x")

		var verr *apperror.ValidationError
		require.ErrorAs(t, v.Err(), &verr)
		verr.FieldErrors["a"] = strings.ToUpper("changed")

		var again *apperror.ValidationError
		require.ErrorAs(t, v.Err(), &again)
		assert.Equal(t, "x", again.FieldErrors["a"])
	})
}
