package middleware

import (
	"context"
	"strings"

	"github.com/labstack/echo/v5"

	"snippetbox/backend-user/internal/apperror"
	"snippetbox/backend-user/internal/model"
)

const UserContextKey = "snippetbox.user"

const bearerPrefix = "Bearer "

type Authenticator interface {
	Authenticate(ctx context.Context, token string) (*model.User, error)
}

func Auth(authenticator Authenticator) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c *echo.Context) error {
			token := bearerToken(c.Request().Header.Get("Authorization"))
			if token == "" {
				return apperror.ErrUnauthenticated
			}

			user, err := authenticator.Authenticate(c.Request().Context(), token)
			if err != nil {
				return err
			}

			c.Set(UserContextKey, user)
			c.Response().Header().Set("Cache-Control", "no-store")
			return next(c)
		}
	}
}

func CurrentUser(c *echo.Context) *model.User {
	user, _ := c.Get(UserContextKey).(*model.User)
	return user
}

func bearerToken(header string) string {
	if len(header) <= len(bearerPrefix) || !strings.EqualFold(header[:len(bearerPrefix)], bearerPrefix) {
		return ""
	}
	return strings.TrimSpace(header[len(bearerPrefix):])
}
