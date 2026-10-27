package middleware

import (
	"context"
	"strings"

	"github.com/labstack/echo/v5"

	"snippetbox/backend-snippet/internal/apperror"
	"snippetbox/backend-snippet/internal/model"
)

const PrincipalContextKey = "snippetbox.principal"

const bearerPrefix = "Bearer "

type Authenticator interface {
	Authenticate(ctx context.Context, token string) (*model.Principal, error)
}

func Auth(authenticator Authenticator) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c *echo.Context) error {
			token := bearerToken(c.Request().Header.Get("Authorization"))
			if token == "" {
				return apperror.ErrUnauthenticated
			}

			principal, err := authenticator.Authenticate(c.Request().Context(), token)
			if err != nil {
				return err
			}

			c.Set(PrincipalContextKey, principal)
			c.Response().Header().Set("Cache-Control", "no-store")
			return next(c)
		}
	}
}

func CurrentPrincipal(c *echo.Context) *model.Principal {
	principal, _ := c.Get(PrincipalContextKey).(*model.Principal)
	return principal
}

func bearerToken(header string) string {
	if len(header) <= len(bearerPrefix) || !strings.EqualFold(header[:len(bearerPrefix)], bearerPrefix) {
		return ""
	}
	return strings.TrimSpace(header[len(bearerPrefix):])
}
