package model_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"snippetbox/backend-user/internal/model"
)

func TestNormalizeEmail(t *testing.T) {
	t.Parallel()

	tests := map[string]string{
		"  Ada@Example.COM ": "ada@example.com",
		"plain@example.com":  "plain@example.com",
		"":                   "",
	}

	for input, want := range tests {
		assert.Equal(t, want, model.NormalizeEmail(input), "input %q", input)
	}
}

func TestTableNames(t *testing.T) {
	t.Parallel()

	assert.Equal(t, "users", model.User{}.TableName())
}
