package dto_test

import (
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"snippetbox/backend-snippet/internal/apperror"
	"snippetbox/backend-snippet/internal/dto"
	"snippetbox/backend-snippet/internal/model"
)

func fieldErrors(t *testing.T, err error) map[string]string {
	t.Helper()
	if err == nil {
		return nil
	}
	var verr *apperror.ValidationError
	require.ErrorAs(t, err, &verr)
	return verr.FieldErrors
}

func TestCreateSnippetRequest_Validate(t *testing.T) {
	t.Parallel()

	valid := dto.CreateSnippetRequest{Title: "t", Content: "c", Expires: 7}

	tests := []struct {
		name   string
		mutate func(*dto.CreateSnippetRequest)
		want   map[string]string
	}{
		{"valid", func(*dto.CreateSnippetRequest) {}, nil},
		{"blank", func(r *dto.CreateSnippetRequest) { *r = dto.CreateSnippetRequest{} }, map[string]string{
			"title": "This field cannot be blank", "content": "This field cannot be blank", "expires": "This field must equal 1, 7 or 365",
		}},
		{"title too long", func(r *dto.CreateSnippetRequest) { r.Title = strings.Repeat("a", 101) }, map[string]string{
			"title": "This field cannot be more than 100 characters long",
		}},
		{"content too long", func(r *dto.CreateSnippetRequest) { r.Content = strings.Repeat("a", 65536) }, map[string]string{
			"content": "This field is too long",
		}},
		{"content at limit", func(r *dto.CreateSnippetRequest) { r.Content = strings.Repeat("a", 65535) }, nil},
		{"expires not permitted", func(r *dto.CreateSnippetRequest) { r.Expires = 30 }, map[string]string{
			"expires": "This field must equal 1, 7 or 365",
		}},
		{"expires 1", func(r *dto.CreateSnippetRequest) { r.Expires = 1 }, nil},
		{"expires 365", func(r *dto.CreateSnippetRequest) { r.Expires = 365 }, nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			req := valid
			tt.mutate(&req)
			assert.Equal(t, tt.want, fieldErrors(t, req.Validate()))
		})
	}
}

func TestNewSnippetListResponse(t *testing.T) {
	t.Parallel()

	t.Run("nil input serializes as empty array", func(t *testing.T) {
		t.Parallel()
		raw, err := json.Marshal(dto.NewSnippetListResponse(nil))
		require.NoError(t, err)
		assert.JSONEq(t, `[]`, string(raw))
	})

	t.Run("maps every snippet in order", func(t *testing.T) {
		t.Parallel()
		now := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
		list := dto.NewSnippetListResponse([]model.Snippet{
			{ID: 2, Title: "b", Created: now, Expires: now.Add(time.Hour)},
			{ID: 1, Title: "a", Created: now, Expires: now.Add(time.Hour)},
		})
		require.Len(t, list, 2)
		assert.Equal(t, uint(2), list[0].ID)
		assert.Equal(t, "a", list[1].Title)
	})
}

func TestNewErrorResponse(t *testing.T) {
	t.Parallel()

	raw, err := json.Marshal(dto.NewErrorResponse(dto.CodeNotFound, "gone"))
	require.NoError(t, err)
	assert.JSONEq(t, `{"error":{"code":"NOT_FOUND","message":"gone"}}`, string(raw))
}
