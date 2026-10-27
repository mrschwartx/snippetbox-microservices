package handler_test

import (
	"errors"
	"fmt"
	"net/http"
	"testing"

	"github.com/labstack/echo/v5"
	"github.com/labstack/echo/v5/echotest"
	"github.com/stretchr/testify/assert"

	"snippetbox/backend-user/internal/apperror"
	"snippetbox/backend-user/internal/handler"
)

func TestErrorHandler_MapsErrorsToEnvelope(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		err    error
		status int
		body   string
	}{
		{
			"validation", apperror.NewFieldValidationError("email", "bad"),
			http.StatusUnprocessableEntity,
			`{"error":{"code":"VALIDATION_ERROR","message":"Validation failed","fields":{"email":"bad"}}}`,
		},
		{
			"validation with non field errors", &apperror.ValidationError{NonFieldErrors: []string{"nope"}},
			http.StatusUnprocessableEntity,
			`{"error":{"code":"VALIDATION_ERROR","message":"Validation failed","non_field":["nope"]}}`,
		},
		{
			"wrapped validation", fmt.Errorf("svc: %w", apperror.NewFieldValidationError("a", "b")),
			http.StatusUnprocessableEntity,
			`{"error":{"code":"VALIDATION_ERROR","message":"Validation failed","fields":{"a":"b"}}}`,
		},
		{
			"not found", apperror.ErrNotFound, http.StatusNotFound,
			`{"error":{"code":"NOT_FOUND","message":"Resource not found"}}`,
		},
		{
			"invalid credentials", apperror.ErrInvalidCredentials, http.StatusUnauthorized,
			`{"error":{"code":"INVALID_CREDENTIALS","message":"Email or password is incorrect"}}`,
		},
		{
			"unauthenticated", fmt.Errorf("ctx: %w", apperror.ErrUnauthenticated), http.StatusUnauthorized,
			`{"error":{"code":"UNAUTHENTICATED","message":"Authentication required"}}`,
		},
		{
			"echo not found", echo.ErrNotFound, http.StatusNotFound,
			`{"error":{"code":"NOT_FOUND","message":"Not Found"}}`,
		},
		{
			"echo method not allowed", echo.ErrMethodNotAllowed, http.StatusMethodNotAllowed,
			`{"error":{"code":"METHOD_NOT_ALLOWED","message":"Method Not Allowed"}}`,
		},
		{
			"echo bad request", echo.NewHTTPError(http.StatusBadRequest, "x"), http.StatusBadRequest,
			`{"error":{"code":"BAD_REQUEST","message":"Bad Request"}}`,
		},
		{
			"other echo status", echo.NewHTTPError(http.StatusTeapot, "x"), http.StatusTeapot,
			`{"error":{"code":"HTTP_ERROR","message":"I'm a teapot"}}`,
		},
		{
			"unknown error", errors.New("db exploded"), http.StatusInternalServerError,
			`{"error":{"code":"INTERNAL_ERROR","message":"Internal server error"}}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			log, _ := newLogger()

			rec := echotest.ContextConfig{}.ServeWithHandler(t, func(*echo.Context) error { return tt.err }, handler.NewErrorHandler(log))

			assert.Equal(t, tt.status, rec.Code)
			assert.Equal(t, echo.MIMEApplicationJSON, rec.Header().Get(echo.HeaderContentType))
			assert.JSONEq(t, tt.body, rec.Body.String())
		})
	}
}

func TestErrorHandler_LogsOnlyServerErrors(t *testing.T) {
	t.Parallel()

	t.Run("5xx is logged with the cause but not leaked to the client", func(t *testing.T) {
		t.Parallel()
		log, buf := newLogger()

		rec := echotest.ContextConfig{}.ServeWithHandler(t, func(*echo.Context) error {
			return errors.New("connection refused to 10.0.0.5")
		}, handler.NewErrorHandler(log))

		assert.Contains(t, buf.String(), "connection refused to 10.0.0.5")
		assert.NotContains(t, rec.Body.String(), "10.0.0.5")
	})

	t.Run("4xx is not logged", func(t *testing.T) {
		t.Parallel()
		log, buf := newLogger()

		echotest.ContextConfig{}.ServeWithHandler(t, func(*echo.Context) error { return apperror.ErrNotFound }, handler.NewErrorHandler(log))

		assert.Empty(t, buf.String())
	})
}

func TestErrorHandler_DoesNotOverwriteCommittedResponse(t *testing.T) {
	t.Parallel()
	log, buf := newLogger()

	rec := echotest.ContextConfig{}.ServeWithHandler(t, func(c *echo.Context) error {
		if err := c.String(http.StatusOK, "partial"); err != nil {
			return err
		}
		return errors.New("late failure")
	}, handler.NewErrorHandler(log))

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "partial", rec.Body.String())
	assert.Contains(t, buf.String(), "late failure")
}

func TestErrorHandler_IgnoresNilError(t *testing.T) {
	t.Parallel()
	log, _ := newLogger()
	c, rec := echotest.ContextConfig{}.ToContextRecorder(t)

	handler.NewErrorHandler(log)(c, nil)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Empty(t, rec.Body.String())
}
