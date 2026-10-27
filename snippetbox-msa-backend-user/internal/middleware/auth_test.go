package middleware_test

import (
	"context"
	"net/http"
	"testing"

	"github.com/labstack/echo/v5"
	"github.com/labstack/echo/v5/echotest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"snippetbox/backend-user/internal/apperror"
	"snippetbox/backend-user/internal/middleware"
	"snippetbox/backend-user/internal/model"
)

type authenticatorFunc func(ctx context.Context, token string) (*model.User, error)

func (f authenticatorFunc) Authenticate(ctx context.Context, token string) (*model.User, error) {
	return f(ctx, token)
}

func TestAuth(t *testing.T) {
	t.Parallel()

	user := &model.User{ID: 9, Email: "ada@example.com"}

	var gotToken string
	authenticator := authenticatorFunc(func(_ context.Context, token string) (*model.User, error) {
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
		assert.Same(t, user, middleware.CurrentUser(c))
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
			assert.Nil(t, middleware.CurrentUser(c))
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

func TestCurrentUser(t *testing.T) {
	t.Parallel()

	c, _ := echotest.ContextConfig{}.ToContextRecorder(t)
	assert.Nil(t, middleware.CurrentUser(c))

	c.Set(middleware.UserContextKey, "not a user")
	assert.Nil(t, middleware.CurrentUser(c))

	user := &model.User{ID: 1}
	c.Set(middleware.UserContextKey, user)
	assert.Same(t, user, middleware.CurrentUser(c))
}
