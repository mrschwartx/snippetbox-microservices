package dto_test

import (
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"snippetbox/backend-user/internal/apperror"
	"snippetbox/backend-user/internal/dto"
	"snippetbox/backend-user/internal/model"
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

func TestSignupRequest_Validate(t *testing.T) {
	t.Parallel()

	valid := dto.SignupRequest{Name: "Ada", Email: "ada@example.com", Password: "pa55word"}

	tests := []struct {
		name   string
		mutate func(*dto.SignupRequest)
		want   map[string]string
	}{
		{"valid", func(*dto.SignupRequest) {}, nil},
		{"blank everything", func(r *dto.SignupRequest) { *r = dto.SignupRequest{} }, map[string]string{
			"name": "This field cannot be blank", "email": "This field cannot be blank", "password": "This field cannot be blank",
		}},
		{"name too long", func(r *dto.SignupRequest) { r.Name = strings.Repeat("a", 256) }, map[string]string{
			"name": "This field cannot be more than 255 characters long",
		}},
		{"name at limit", func(r *dto.SignupRequest) { r.Name = strings.Repeat("a", 255) }, nil},
		{"invalid email", func(r *dto.SignupRequest) { r.Email = "nope" }, map[string]string{
			"email": "This field must be a valid email address",
		}},
		{"email too long", func(r *dto.SignupRequest) { r.Email = strings.Repeat("a", 250) + "@b.com" }, map[string]string{
			"email": "This field cannot be more than 255 characters long",
		}},
		{"password too short", func(r *dto.SignupRequest) { r.Password = "1234567" }, map[string]string{
			"password": "This field must be at least 8 characters long",
		}},
		{"password over 72 bytes", func(r *dto.SignupRequest) { r.Password = strings.Repeat("a", 73) }, map[string]string{
			"password": "This field cannot be more than 72 bytes long",
		}},
		{"password at 72 bytes", func(r *dto.SignupRequest) { r.Password = strings.Repeat("a", 72) }, nil},
		{"multibyte password over 72 bytes", func(r *dto.SignupRequest) { r.Password = strings.Repeat("é", 37) }, map[string]string{
			"password": "This field cannot be more than 72 bytes long",
		}},
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

func TestLoginRequest_Validate(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		req  dto.LoginRequest
		want map[string]string
	}{
		{"valid", dto.LoginRequest{Email: "a@b.co", Password: "x"}, nil},
		{"blank", dto.LoginRequest{}, map[string]string{
			"email": "This field cannot be blank", "password": "This field cannot be blank",
		}},
		{"invalid email", dto.LoginRequest{Email: "a", Password: "x"}, map[string]string{
			"email": "This field must be a valid email address",
		}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.want, fieldErrors(t, tt.req.Validate()))
		})
	}
}

func TestNewUserResponse_OmitsPasswordHash(t *testing.T) {
	t.Parallel()

	created := time.Date(2026, 1, 2, 3, 4, 5, 0, time.FixedZone("WIB", 7*3600))
	user := &model.User{ID: 4, Name: "Ada", Email: "ada@example.com", HashedPassword: "secret", Created: created}

	raw, err := json.Marshal(dto.NewUserResponse(user))
	require.NoError(t, err)

	assert.JSONEq(t, `{"id":4,"name":"Ada","email":"ada@example.com","created":"2026-01-01T20:04:05Z"}`, string(raw))
	assert.NotContains(t, string(raw), "secret")
}

func TestNewLoginResponse(t *testing.T) {
	t.Parallel()

	expires := time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC)
	resp := dto.NewLoginResponse(model.Token{Value: "jwt", ExpiresAt: expires}, &model.User{ID: 1})

	assert.Equal(t, "jwt", resp.Token)
	assert.Equal(t, "Bearer", resp.TokenType)
	assert.Equal(t, expires, resp.ExpiresAt)
	assert.Equal(t, uint(1), resp.User.ID)
}

func TestNewErrorResponse(t *testing.T) {
	t.Parallel()

	raw, err := json.Marshal(dto.NewErrorResponse(dto.CodeNotFound, "gone"))
	require.NoError(t, err)
	assert.JSONEq(t, `{"error":{"code":"NOT_FOUND","message":"gone"}}`, string(raw))
}
