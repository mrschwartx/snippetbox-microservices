package handler

import (
	"fmt"
	"net/http"

	"github.com/labstack/echo/v5"

	"snippetbox/backend-snippet/internal/dto"
	"snippetbox/backend-snippet/internal/middleware"
)

type SnippetHandler struct {
	snippets SnippetService
}

func NewSnippetHandler(snippets SnippetService) *SnippetHandler {
	return &SnippetHandler{snippets: snippets}
}

func (h *SnippetHandler) List(c *echo.Context) error {
	snippets, err := h.snippets.ListLatest(c.Request().Context())
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, dto.NewSnippetListResponse(snippets))
}

func (h *SnippetHandler) Get(c *echo.Context) error {
	id, err := parseID(c.Param("id"))
	if err != nil {
		return err
	}

	snippet, err := h.snippets.Get(c.Request().Context(), id)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, dto.NewSnippetResponse(snippet))
}

func (h *SnippetHandler) Create(c *echo.Context) error {
	var req dto.CreateSnippetRequest
	if err := bindJSON(c, &req); err != nil {
		return err
	}

	snippet, err := h.snippets.Create(c.Request().Context(), middleware.CurrentPrincipal(c), req)
	if err != nil {
		return err
	}

	c.Response().Header().Set("Location", fmt.Sprintf("%s/snippets/%d", APIPrefix, snippet.ID))
	return c.JSON(http.StatusCreated, dto.NewSnippetResponse(snippet))
}
