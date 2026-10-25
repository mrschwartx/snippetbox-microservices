package handler_test

import (
	"net/http"
	"testing"

	"github.com/labstack/echo/v5/echotest"
	"github.com/stretchr/testify/assert"

	"snippetbox/backend-snippet/internal/handler"
)

func TestSystemHandler(t *testing.T) {
	t.Parallel()

	h := handler.NewSystemHandler([]byte(`{"openapi":"3.0.3"}`), "<html>swagger</html>")

	t.Run("ping", func(t *testing.T) {
		t.Parallel()
		status, body, _ := serveHandler(t, h.Ping, echotest.ContextConfig{})

		assert.Equal(t, http.StatusOK, status)
		assert.JSONEq(t, `{"status":"ok"}`, body)
	})

	t.Run("swagger without trailing slash redirects", func(t *testing.T) {
		t.Parallel()
		status, _, headers := serveHandler(t, h.SwaggerRedirect, echotest.ContextConfig{})

		assert.Equal(t, http.StatusMovedPermanently, status)
		assert.Equal(t, "/swagger/", headers.Get("Location"))
	})

	t.Run("swagger UI serves the embedded page", func(t *testing.T) {
		t.Parallel()
		status, body, headers := serveHandler(t, h.SwaggerUI, echotest.ContextConfig{})

		assert.Equal(t, http.StatusOK, status)
		assert.Contains(t, headers.Get("Content-Type"), "text/html")
		assert.Equal(t, "<html>swagger</html>", body)
	})

	t.Run("openapi spec is served as JSON", func(t *testing.T) {
		t.Parallel()
		status, body, headers := serveHandler(t, h.OpenAPI, echotest.ContextConfig{})

		assert.Equal(t, http.StatusOK, status)
		assert.Contains(t, headers.Get("Content-Type"), "application/json")
		assert.JSONEq(t, `{"openapi":"3.0.3"}`, body)
	})
}
