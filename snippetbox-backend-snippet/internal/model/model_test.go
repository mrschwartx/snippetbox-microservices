package model_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"snippetbox/backend-snippet/internal/model"
)

var now = time.Date(2026, 10, 15, 8, 0, 0, 0, time.UTC)

func TestNewSnippet(t *testing.T) {
	t.Parallel()

	local := time.FixedZone("WIB", 7*60*60)

	for _, days := range model.PermittedExpires {
		snippet := model.NewSnippet("title", "body", days, now.In(local))

		assert.Equal(t, "title", snippet.Title)
		assert.Equal(t, "body", snippet.Content)
		assert.Equal(t, time.UTC, snippet.Created.Location())
		assert.True(t, now.Equal(snippet.Created))
		assert.Equal(t, now.Add(time.Duration(days)*24*time.Hour), snippet.Expires)
	}
}

func TestSnippet_IsExpired(t *testing.T) {
	t.Parallel()

	snippet := model.NewSnippet("t", "c", 1, now)

	tests := []struct {
		name string
		at   time.Time
		want bool
	}{
		{"just created", now, false},
		{"one second before expiry", snippet.Expires.Add(-time.Second), false},
		{"exactly at expiry", snippet.Expires, true},
		{"after expiry", snippet.Expires.Add(time.Second), true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.want, snippet.IsExpired(tt.at))
		})
	}
}

func TestTableNames(t *testing.T) {
	t.Parallel()

	assert.Equal(t, "snippets", model.Snippet{}.TableName())
}
