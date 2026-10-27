package middleware_test

import (
	"context"
	"net/http"
	"testing"

	"github.com/labstack/echo/v5"
	"github.com/labstack/echo/v5/echotest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"snippetbox/backend-snippet/internal/apperror"
	"snippetbox/backend-snippet/internal/middleware"
	"snippetbox/backend-snippet/internal/model"
)

type authenticatorFunc func(ctx context.Context, token string) (*model.Principal, error)

func (f authenticatorFunc) Authenticate(ctx context.Context, token string) (*model.Principal, error) {
	return f(ctx, token)
}

func TestAuth(t *testing.T) {
	t.Parallel()

	user := &model.Principal{ID: 9}

	var gotToken string
	authenticator := authenticatorFunc(func(_ context.Context, token string) (*model.Principal, error) {
		gotToken = token
		switch token {
		case "good":
			return user, nil
		case "broken":
			return nil, assert.AnError
		default:
			return nil, apperror.ErrUnauthenticated
		}
	})

	run := func(t *testing.T, authorization string) (*echo.Context, error, bool) {
		t.Helper()
		headers := http.Header{}
		if authorization != "" {
			headers.Set("Authorization", authorization)
		}
		c, _ := echotest.ContextConfig{Headers: headers}.ToContextRecorder(t)

		called := false
		err := middleware.Auth(authenticator)(func(*echo.Context) error {
			called = true
			return nil
		})(c)
		return c, err, called
	}

	t.Run("valid token stores the user and disables caching", func(t *testing.T) {
		c, err, called := run(t, "Bearer good")

		require.NoError(t, err)
		assert.True(t, called)
		assert.Same(t, user, middleware.CurrentPrincipal(c))
		assert.Equal(t, "no-store", c.Response().Header().Get("Cache-Control"))
		assert.Equal(t, "good", gotToken)
	})

	t.Run("scheme is case insensitive and the token is trimmed", func(t *testing.T) {
		_, err, called := run(t, "bearer   good  ")

		require.NoError(t, err)
		assert.True(t, called)
		assert.Equal(t, "good", gotToken)
	})

	t.Run("missing or malformed credentials never reach the authenticator", func(t *testing.T) {
		for _, header := range []string{"", "Bearer", "Bearer ", "Bearer    ", "Basic good", "good", "Token good"} {
			gotToken = "untouched"

			c, err, called := run(t, header)

			assert.ErrorIs(t, err, apperror.ErrUnauthenticated, "header %q", header)
			assert.False(t, called, "header %q", header)
			assert.Equal(t, "untouched", gotToken, "header %q", header)
			assert.Nil(t, middleware.CurrentPrincipal(c))
		}
	})

	t.Run("authenticator errors are returned unchanged", func(t *testing.T) {
		_, err, called := run(t, "Bearer broken")

		assert.ErrorIs(t, err, assert.AnError)
		assert.False(t, called)

		_, err, _ = run(t, "Bearer nope")
		assert.ErrorIs(t, err, apperror.ErrUnauthenticated)
	})
}

func TestCurrentPrincipal(t *testing.T) {
	t.Parallel()

	c, _ := echotest.ContextConfig{}.ToContextRecorder(t)
	assert.Nil(t, middleware.CurrentPrincipal(c))

	c.Set(middleware.PrincipalContextKey, "not a principal")
	assert.Nil(t, middleware.CurrentPrincipal(c))

	user := &model.Principal{ID: 1}
	c.Set(middleware.PrincipalContextKey, user)
	assert.Same(t, user, middleware.CurrentPrincipal(c))
}
