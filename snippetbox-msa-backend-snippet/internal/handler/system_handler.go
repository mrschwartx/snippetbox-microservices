package handler

import (
	"net/http"

	"github.com/labstack/echo/v5"

	"snippetbox/backend-snippet/internal/dto"
)

type SystemHandler struct {
	openAPISpec []byte
	swaggerPage string
}

func NewSystemHandler(openAPISpec []byte, swaggerPage string) *SystemHandler {
	return &SystemHandler{openAPISpec: openAPISpec, swaggerPage: swaggerPage}
}

func (h *SystemHandler) Ping(c *echo.Context) error {
	return c.JSON(http.StatusOK, dto.StatusResponse{Status: "ok"})
}

func (h *SystemHandler) SwaggerRedirect(c *echo.Context) error {
	return c.Redirect(http.StatusMovedPermanently, "/swagger/")
}

func (h *SystemHandler) SwaggerUI(c *echo.Context) error {
	return c.HTML(http.StatusOK, h.swaggerPage)
}

func (h *SystemHandler) OpenAPI(c *echo.Context) error {
	return c.Blob(http.StatusOK, echo.MIMEApplicationJSON, h.openAPISpec)
}
