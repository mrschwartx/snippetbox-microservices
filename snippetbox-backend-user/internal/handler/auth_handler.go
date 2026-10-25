package handler

import (
	"net/http"

	"github.com/labstack/echo/v5"

	"snippetbox/backend-user/internal/apperror"
	"snippetbox/backend-user/internal/dto"
	"snippetbox/backend-user/internal/middleware"
)

type AuthHandler struct {
	auth AuthService
}

func NewAuthHandler(auth AuthService) *AuthHandler {
	return &AuthHandler{auth: auth}
}

func (h *AuthHandler) Signup(c *echo.Context) error {
	var req dto.SignupRequest
	if err := bindJSON(c, &req); err != nil {
		return err
	}

	user, err := h.auth.Signup(c.Request().Context(), req)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusCreated, dto.NewUserResponse(user))
}

func (h *AuthHandler) Login(c *echo.Context) error {
	var req dto.LoginRequest
	if err := bindJSON(c, &req); err != nil {
		return err
	}

	token, user, err := h.auth.Login(c.Request().Context(), req)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, dto.NewLoginResponse(token, user))
}

func (h *AuthHandler) Logout(c *echo.Context) error {
	return c.NoContent(http.StatusNoContent)
}

func (h *AuthHandler) Me(c *echo.Context) error {
	user := middleware.CurrentUser(c)
	if user == nil {
		return apperror.ErrUnauthenticated
	}

	return c.JSON(http.StatusOK, dto.NewUserResponse(user))
}
