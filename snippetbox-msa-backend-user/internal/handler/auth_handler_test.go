package handler_test

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/labstack/echo/v5"
	"github.com/labstack/echo/v5/echotest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"snippetbox/backend-user/internal/apperror"
	"snippetbox/backend-user/internal/dto"
	"snippetbox/backend-user/internal/handler"
	"snippetbox/backend-user/internal/middleware"
	"snippetbox/backend-user/internal/model"
)

var created = time.Date(2026, 10, 15, 8, 0, 0, 0, time.UTC)

func serveHandler(t *testing.T, h echo.HandlerFunc, cfg echotest.ContextConfig) (int, string, http.Header) {
	t.Helper()
	log, _ := newLogger()
	rec := cfg.ServeWithHandler(t, h, handler.NewErrorHandler(log))
	return rec.Code, rec.Body.String(), rec.Header()
}

func TestAuthHandler_Signup(t *testing.T) {
	t.Parallel()

	t.Run("creates the user and never exposes the password hash", func(t *testing.T) {
		t.Parallel()
		var got dto.SignupRequest
		h := handler.NewAuthHandler(authServiceFake{signup: func(_ context.Context, req dto.SignupRequest) (*model.User, error) {
			got = req
			return &model.User{ID: 3, Name: "Ada", Email: "ada@example.com", HashedPassword: "secret", Created: created}, nil
		}})

		status, body, _ := serveHandler(t, h.Signup, echotest.ContextConfig{
			JSONBody: []byte(`{"name":"Ada","email":"ada@example.com","password":"pa55word"}`),
		})

		assert.Equal(t, http.StatusCreated, status)
		assert.JSONEq(t, `{"id":3,"name":"Ada","email":"ada@example.com","created":"2026-10-15T08:00:00Z"}`, body)
		assert.Equal(t, dto.SignupRequest{Name: "Ada", Email: "ada@example.com", Password: "pa55word"}, got)
	})

	t.Run("rejects bodies that are not strict JSON", func(t *testing.T) {
		t.Parallel()
		h := handler.NewAuthHandler(authServiceFake{signup: func(context.Context, dto.SignupRequest) (*model.User, error) {
			t.Fatal("service must not be called")
			return nil, nil
		}})

		for name, body := range map[string]string{
			"empty":            ``,
			"not json":         `nope`,
			"unknown field":    `{"name":"a","email":"a@b.co","password":"12345678","admin":true}`,
			"trailing data":    `{"name":"a"} {"name":"b"}`,
			"wrong field type": `{"name":1}`,
		} {
			status, resp, _ := serveHandler(t, h.Signup, echotest.ContextConfig{JSONBody: []byte(body)})

			assert.Equal(t, http.StatusBadRequest, status, name)
			assert.Contains(t, resp, `"BAD_REQUEST"`, name)
		}
	})

	t.Run("surfaces service errors through the error handler", func(t *testing.T) {
		t.Parallel()
		h := handler.NewAuthHandler(authServiceFake{signup: func(context.Context, dto.SignupRequest) (*model.User, error) {
			return nil, apperror.NewFieldValidationError("email", "taken")
		}})

		status, body, _ := serveHandler(t, h.Signup, echotest.ContextConfig{JSONBody: []byte(`{}`)})

		assert.Equal(t, http.StatusUnprocessableEntity, status)
		assert.Contains(t, body, `"email":"taken"`)
	})
}

func TestAuthHandler_Login(t *testing.T) {
	t.Parallel()

	t.Run("returns a bearer token with the user", func(t *testing.T) {
		t.Parallel()
		h := handler.NewAuthHandler(authServiceFake{login: func(_ context.Context, req dto.LoginRequest) (model.Token, *model.User, error) {
			assert.Equal(t, dto.LoginRequest{Email: "ada@example.com", Password: "pa55word"}, req)
			return model.Token{Value: "jwt", ExpiresAt: created.Add(time.Hour)}, &model.User{ID: 3, Name: "Ada", Email: "ada@example.com", Created: created}, nil
		}})

		status, body, _ := serveHandler(t, h.Login, echotest.ContextConfig{
			JSONBody: []byte(`{"email":"ada@example.com","password":"pa55word"}`),
		})

		assert.Equal(t, http.StatusOK, status)
		assert.JSONEq(t, `{
			"token":"jwt","token_type":"Bearer","expires_at":"2026-10-15T09:00:00Z",
			"user":{"id":3,"name":"Ada","email":"ada@example.com","created":"2026-10-15T08:00:00Z"}
		}`, body)
	})

	t.Run("maps invalid credentials to 401", func(t *testing.T) {
		t.Parallel()
		h := handler.NewAuthHandler(authServiceFake{login: func(context.Context, dto.LoginRequest) (model.Token, *model.User, error) {
			return model.Token{}, nil, apperror.ErrInvalidCredentials
		}})

		status, body, _ := serveHandler(t, h.Login, echotest.ContextConfig{JSONBody: []byte(`{"email":"a@b.co","password":"x"}`)})

		assert.Equal(t, http.StatusUnauthorized, status)
		assert.Contains(t, body, `"INVALID_CREDENTIALS"`)
	})

	t.Run("rejects a malformed body", func(t *testing.T) {
		t.Parallel()
		h := handler.NewAuthHandler(authServiceFake{})

		status, _, _ := serveHandler(t, h.Login, echotest.ContextConfig{JSONBody: []byte(`[`)})

		assert.Equal(t, http.StatusBadRequest, status)
	})
}

func TestAuthHandler_Logout(t *testing.T) {
	t.Parallel()

	status, body, _ := serveHandler(t, handler.NewAuthHandler(authServiceFake{}).Logout, echotest.ContextConfig{})

	assert.Equal(t, http.StatusNoContent, status)
	assert.Empty(t, body)
}

func TestAuthHandler_Me(t *testing.T) {
	t.Parallel()
	h := handler.NewAuthHandler(authServiceFake{})

	t.Run("returns the authenticated user", func(t *testing.T) {
		t.Parallel()
		c, rec := echotest.ContextConfig{}.ToContextRecorder(t)
		c.Set(middleware.UserContextKey, &model.User{ID: 5, Name: "Ada", Email: "ada@example.com", Created: created})

		require.NoError(t, h.Me(c))

		assert.Equal(t, http.StatusOK, rec.Code)
		assert.JSONEq(t, `{"id":5,"name":"Ada","email":"ada@example.com","created":"2026-10-15T08:00:00Z"}`, rec.Body.String())
	})

	t.Run("is unauthenticated when the route was not guarded", func(t *testing.T) {
		t.Parallel()
		status, body, _ := serveHandler(t, h.Me, echotest.ContextConfig{})

		assert.Equal(t, http.StatusUnauthorized, status)
		assert.Contains(t, body, `"UNAUTHENTICATED"`)
	})
}
