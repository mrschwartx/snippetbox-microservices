package router_test

import (
	"bytes"
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"snippetbox/backend-snippet/internal/apperror"
	"snippetbox/backend-snippet/internal/dto"
	"snippetbox/backend-snippet/internal/handler"
	"snippetbox/backend-snippet/internal/model"
	"snippetbox/backend-snippet/internal/router"
)

var created = time.Date(2026, 10, 15, 8, 0, 0, 0, time.UTC)

type snippetFake struct{ panicOnList bool }

func (snippetFake) Create(context.Context, *model.Principal, dto.CreateSnippetRequest) (*model.Snippet, error) {
	return &model.Snippet{ID: 5, Created: created, Expires: created}, nil
}

func (snippetFake) Get(_ context.Context, id uint) (*model.Snippet, error) {
	if id != 5 {
		return nil, apperror.ErrNotFound
	}
	return &model.Snippet{ID: 5, Title: "t", Content: "c", Created: created, Expires: created}, nil
}

func (f snippetFake) ListLatest(context.Context) ([]model.Snippet, error) {
	if f.panicOnList {
		panic("list exploded")
	}
	return nil, nil
}

type authenticatorFake struct{}

func (authenticatorFake) Authenticate(_ context.Context, token string) (*model.Principal, error) {
	if token != "good" {
		return nil, apperror.ErrUnauthenticated
	}
	return &model.Principal{ID: 1}, nil
}

type setup struct {
	cfg  router.Config
	logs *bytes.Buffer
}

func newSetup() setup {
	var logs bytes.Buffer
	return setup{
		logs: &logs,
		cfg: router.Config{
			Snippets:       handler.NewSnippetHandler(snippetFake{}),
			System:         handler.NewSystemHandler([]byte(`{"openapi":"3.0.3"}`), "<html>ui</html>"),
			Authenticator:  authenticatorFake{},
			Logger:         slog.New(slog.NewJSONHandler(&logs, nil)),
			AllowedOrigins: []string{"https://app.test"},
		},
	}
}

func do(h http.Handler, method, target, body string, headers map[string]string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(method, target, strings.NewReader(body))
	for key, value := range headers {
		req.Header.Set(key, value)
	}
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	return rec
}

var bearer = map[string]string{"Authorization": "Bearer good"}

func TestRouter_Routes(t *testing.T) {
	t.Parallel()
	h := router.New(newSetup().cfg)

	tests := []struct {
		name    string
		method  string
		target  string
		body    string
		headers map[string]string
		status  int
	}{
		{"ping", http.MethodGet, "/ping", "", nil, http.StatusOK},
		{"swagger redirect", http.MethodGet, "/swagger", "", nil, http.StatusMovedPermanently},
		{"swagger ui", http.MethodGet, "/swagger/", "", nil, http.StatusOK},
		{"openapi", http.MethodGet, "/swagger/openapi.json", "", nil, http.StatusOK},
		{"list snippets is public", http.MethodGet, "/api/v1/snippets", "", nil, http.StatusOK},
		{"get snippet is public", http.MethodGet, "/api/v1/snippets/5", "", nil, http.StatusOK},
		{"get missing snippet", http.MethodGet, "/api/v1/snippets/6", "", nil, http.StatusNotFound},
		{"create snippet with token", http.MethodPost, "/api/v1/snippets", `{}`, bearer, http.StatusCreated},
		{"unknown route", http.MethodGet, "/nope", "", nil, http.StatusNotFound},
		{"unknown api route", http.MethodGet, "/api/v1/nope", "", nil, http.StatusNotFound},
		{"wrong method", http.MethodDelete, "/api/v1/snippets", "", nil, http.StatusMethodNotAllowed},
		{"auth belongs to the user service", http.MethodPost, "/api/v1/auth/login", `{}`, nil, http.StatusNotFound},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			rec := do(h, tt.method, tt.target, tt.body, tt.headers)

			assert.Equal(t, tt.status, rec.Code)
		})
	}
}

func TestRouter_ProtectedRoutesRequireAuthentication(t *testing.T) {
	t.Parallel()
	h := router.New(newSetup().cfg)

	protected := []struct{ method, target string }{
		{http.MethodPost, "/api/v1/snippets"},
	}
	credentials := []map[string]string{
		nil,
		{"Authorization": "Bearer wrong"},
		{"Authorization": "Basic good"},
	}

	for _, route := range protected {
		for _, headers := range credentials {
			rec := do(h, route.method, route.target, `{}`, headers)

			assert.Equal(t, http.StatusUnauthorized, rec.Code, "%s %s %v", route.method, route.target, headers)
			assert.JSONEq(t, `{"error":{"code":"UNAUTHENTICATED","message":"Authentication required"}}`, rec.Body.String())
		}
	}
}

func TestRouter_ErrorsUseTheJSONEnvelope(t *testing.T) {
	t.Parallel()
	h := router.New(newSetup().cfg)

	rec := do(h, http.MethodGet, "/missing", "", nil)

	assert.Equal(t, "application/json", rec.Header().Get("Content-Type"))
	assert.JSONEq(t, `{"error":{"code":"NOT_FOUND","message":"Not Found"}}`, rec.Body.String())
}

func TestRouter_MiddlewareChain(t *testing.T) {
	t.Parallel()

	t.Run("every response carries a request id and security headers", func(t *testing.T) {
		t.Parallel()
		h := router.New(newSetup().cfg)

		rec := do(h, http.MethodGet, "/ping", "", map[string]string{"X-Request-ID": "trace-1"})

		assert.Equal(t, "trace-1", rec.Header().Get("X-Request-ID"))
		assert.Equal(t, "nosniff", rec.Header().Get("X-Content-Type-Options"))
		assert.NotEmpty(t, rec.Header().Get("Content-Security-Policy"))
	})

	t.Run("swagger is exempt from the CSP", func(t *testing.T) {
		t.Parallel()
		h := router.New(newSetup().cfg)

		rec := do(h, http.MethodGet, "/swagger/", "", nil)

		assert.Empty(t, rec.Header().Get("Content-Security-Policy"))
	})

	t.Run("preflight requests are answered by CORS before routing", func(t *testing.T) {
		t.Parallel()
		h := router.New(newSetup().cfg)

		rec := do(h, http.MethodOptions, "/api/v1/snippets", "", map[string]string{
			"Origin":                        "https://app.test",
			"Access-Control-Request-Method": "POST",
		})

		assert.Equal(t, http.StatusNoContent, rec.Code)
		assert.Equal(t, "https://app.test", rec.Header().Get("Access-Control-Allow-Origin"))
	})

	t.Run("oversized bodies are rejected with 413", func(t *testing.T) {
		t.Parallel()
		s := newSetup()
		s.cfg.MaxBodyBytes = 32
		h := router.New(s.cfg)

		rec := do(h, http.MethodPost, "/api/v1/snippets", `{"title":"`+strings.Repeat("a", 64)+`"}`, bearer)

		assert.Equal(t, http.StatusRequestEntityTooLarge, rec.Code)
		assert.Contains(t, rec.Body.String(), `"PAYLOAD_TOO_LARGE"`)
	})

	t.Run("the default body limit is 1 MiB", func(t *testing.T) {
		t.Parallel()
		cfg := newSetup().cfg
		cfg.Logger = nil
		h := router.New(cfg)

		within := do(h, http.MethodPost, "/api/v1/snippets", `{"title":"`+strings.Repeat("a", 1<<19)+`"}`, bearer)
		over := do(h, http.MethodPost, "/api/v1/snippets", `{"title":"`+strings.Repeat("a", 1<<20)+`"}`, bearer)

		assert.Equal(t, http.StatusCreated, within.Code)
		assert.Equal(t, http.StatusRequestEntityTooLarge, over.Code)
	})

	t.Run("a panicking handler becomes a logged 500 and the server keeps serving", func(t *testing.T) {
		t.Parallel()
		s := newSetup()
		s.cfg.Snippets = handler.NewSnippetHandler(snippetFake{panicOnList: true})
		h := router.New(s.cfg)

		rec := do(h, http.MethodGet, "/api/v1/snippets", "", nil)

		assert.Equal(t, http.StatusInternalServerError, rec.Code)
		assert.JSONEq(t, `{"error":{"code":"INTERNAL_ERROR","message":"Internal server error"}}`, rec.Body.String())
		assert.Contains(t, s.logs.String(), "list exploded")
		assert.Equal(t, http.StatusOK, do(h, http.MethodGet, "/ping", "", nil).Code)
	})

	t.Run("panicked requests are still access logged as 500", func(t *testing.T) {
		t.Parallel()
		s := newSetup()
		s.cfg.Snippets = handler.NewSnippetHandler(snippetFake{panicOnList: true})

		do(router.New(s.cfg), http.MethodGet, "/api/v1/snippets", "", nil)

		var found bool
		for _, line := range strings.Split(strings.TrimSpace(s.logs.String()), "\n") {
			var entry map[string]any
			require.NoError(t, json.Unmarshal([]byte(line), &entry))
			if entry["msg"] == "request" {
				found = true
				assert.EqualValues(t, http.StatusInternalServerError, entry["status"])
			}
		}
		assert.True(t, found, "no access log entry written")
	})
}
