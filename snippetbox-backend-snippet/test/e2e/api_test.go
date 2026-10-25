//go:build e2e

package e2e_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"snippetbox/backend-snippet/internal/app"
	"snippetbox/backend-snippet/internal/config"
	"snippetbox/backend-snippet/internal/testutil"
)

func TestE2E_HealthAndDocumentation(t *testing.T) {
	a := newAPI(t)

	ping := a.call(http.MethodGet, "/ping", "", nil)
	assert.Equal(t, http.StatusOK, ping.status)
	assert.Equal(t, map[string]any{"status": "ok"}, ping.json(t))

	redirect := a.call(http.MethodGet, "/swagger", "", nil)
	assert.Equal(t, http.StatusMovedPermanently, redirect.status)
	assert.Equal(t, "/swagger/", redirect.headers.Get("Location"))

	ui := a.call(http.MethodGet, "/swagger/", "", nil)
	assert.Equal(t, http.StatusOK, ui.status)
	assert.Contains(t, ui.headers.Get("Content-Type"), "text/html")
	assert.Empty(t, ui.headers.Get("Content-Security-Policy"))

	spec := a.call(http.MethodGet, "/swagger/openapi.json", "", nil)
	assert.Equal(t, http.StatusOK, spec.status)
	assert.Contains(t, spec.json(t), "paths")
}

func TestE2E_SnippetJourney(t *testing.T) {
	a := newAPI(t)
	token := a.issueToken(1)

	created := a.createSnippet(token, "Hello", 7)
	require.Equal(t, http.StatusCreated, created.status, string(created.body))
	snippet := created.json(t)
	location := created.headers.Get("Location")
	assert.Equal(t, fmt.Sprintf("/api/v1/snippets/%.0f", snippet["id"]), location)
	assert.Equal(t, "2026-10-15T08:00:00Z", snippet["created"])
	assert.Equal(t, "2026-10-22T08:00:00Z", snippet["expires"])
	assert.Equal(t, "no-store", created.headers.Get("Cache-Control"))

	fetched := a.call(http.MethodGet, location, "", nil)
	assert.Equal(t, http.StatusOK, fetched.status)
	assert.Equal(t, snippet, fetched.json(t))

	list := a.call(http.MethodGet, "/api/v1/snippets", "", nil).jsonList(t)
	require.Len(t, list, 1)
	assert.Equal(t, snippet["id"], list[0]["id"])
}

func TestE2E_Authentication(t *testing.T) {
	a := newAPI(t)
	token := a.issueToken(1)

	for name, bad := range map[string]string{
		"missing":       "",
		"garbage":       "garbage",
		"tampered":      token + "x",
		"wrong secret":  testutil.IssueToken(t, "another-secret-another-secret-0000", 1, a.clock.Now(), time.Hour),
		"already stale": testutil.IssueToken(t, jwtSecret, 1, a.clock.Now().Add(-2*time.Hour), time.Hour),
	} {
		resp := a.call(http.MethodPost, "/api/v1/snippets", bad, map[string]any{})

		assert.Equal(t, http.StatusUnauthorized, resp.status, name)
		assert.Equal(t, "UNAUTHENTICATED", resp.json(t)["error"].(map[string]any)["code"], name)
	}

	notBearer := a.send(http.MethodPost, "/api/v1/snippets", map[string]string{"Authorization": "Basic " + token}, nil)
	assert.Equal(t, http.StatusUnauthorized, notBearer.status)

	assert.Equal(t, http.StatusCreated, a.createSnippet(token, "fresh", 1).status)

	a.clock.Advance(11*time.Hour + 59*time.Minute)
	assert.Equal(t, http.StatusCreated, a.createSnippet(token, "almost late", 1).status)

	a.clock.Advance(2 * time.Minute)
	assert.Equal(t, http.StatusUnauthorized, a.createSnippet(token, "late", 1).status)

	assert.Equal(t, http.StatusOK, a.call(http.MethodGet, "/api/v1/snippets", "", nil).status)
}

func TestE2E_RequestValidation(t *testing.T) {
	a := newAPI(t)
	token := a.issueToken(1)

	t.Run("snippet field errors", func(t *testing.T) {
		tests := []struct {
			name, body, field string
		}{
			{"blank title", `{"title":"","content":"c","expires":7}`, "title"},
			{"long title", `{"title":"` + strings.Repeat("a", 101) + `","content":"c","expires":7}`, "title"},
			{"blank content", `{"title":"t","content":"","expires":7}`, "content"},
			{"unsupported expiry", `{"title":"t","content":"c","expires":30}`, "expires"},
			{"missing expiry", `{"title":"t","content":"c"}`, "expires"},
		}
		for _, tt := range tests {
			resp := a.raw(http.MethodPost, "/api/v1/snippets", token, tt.body)

			assert.Equal(t, http.StatusUnprocessableEntity, resp.status, tt.name)
			fields := resp.json(t)["error"].(map[string]any)["fields"].(map[string]any)
			assert.Contains(t, fields, tt.field, tt.name)
		}
	})

	t.Run("malformed bodies", func(t *testing.T) {
		for name, body := range map[string]string{
			"not json":       `hello`,
			"empty":          ``,
			"unknown field":  `{"title":"t","content":"c","expires":7,"admin":true}`,
			"wrong type":     `{"title":"t","content":"c","expires":"7"}`,
			"trailing value": `{"title":"t","content":"c","expires":7}{}`,
		} {
			resp := a.raw(http.MethodPost, "/api/v1/snippets", token, body)

			assert.Equal(t, http.StatusBadRequest, resp.status, name)
			assert.Equal(t, "BAD_REQUEST", resp.json(t)["error"].(map[string]any)["code"], name)
		}
	})

	t.Run("bodies over 1 MiB", func(t *testing.T) {
		huge := `{"title":"t","content":"` + strings.Repeat("a", 1<<20) + `","expires":7}`

		resp := a.raw(http.MethodPost, "/api/v1/snippets", token, huge)

		assert.Equal(t, http.StatusRequestEntityTooLarge, resp.status)
		assert.Equal(t, "PAYLOAD_TOO_LARGE", resp.json(t)["error"].(map[string]any)["code"])
	})

	t.Run("content just under the field limit is accepted", func(t *testing.T) {
		body := `{"title":"t","content":"` + strings.Repeat("a", 65535) + `","expires":1}`

		assert.Equal(t, http.StatusCreated, a.raw(http.MethodPost, "/api/v1/snippets", token, body).status)
	})
}

func TestE2E_SnippetVisibility(t *testing.T) {
	a := newAPI(t)
	token := a.issueToken(1)

	t.Run("ids that cannot exist are plain 404s, never server errors", func(t *testing.T) {
		for _, id := range []string{"0", "abc", "-1", "1.5", "99999999999999999999", "9223372036854775807", "9223372036854775808"} {
			resp := a.call(http.MethodGet, "/api/v1/snippets/"+id, "", nil)

			assert.Equal(t, http.StatusNotFound, resp.status, "id %s", id)
			assert.Equal(t, "NOT_FOUND", resp.json(t)["error"].(map[string]any)["code"], "id %s", id)
		}
	})

	t.Run("an empty store lists as an empty array", func(t *testing.T) {
		resp := a.call(http.MethodGet, "/api/v1/snippets", "", nil)

		assert.Equal(t, http.StatusOK, resp.status)
		assert.JSONEq(t, `[]`, string(resp.body))
	})

	t.Run("snippets expire on schedule", func(t *testing.T) {
		short := a.createSnippet(token, "short", 1).json(t)
		long := a.createSnippet(token, "long", 365).json(t)
		shortPath := fmt.Sprintf("/api/v1/snippets/%.0f", short["id"])

		a.clock.Advance(24*time.Hour - time.Second)
		assert.Equal(t, http.StatusOK, a.call(http.MethodGet, shortPath, "", nil).status)
		assert.Len(t, a.call(http.MethodGet, "/api/v1/snippets", "", nil).jsonList(t), 2)

		a.clock.Advance(time.Second)
		assert.Equal(t, http.StatusNotFound, a.call(http.MethodGet, shortPath, "", nil).status)
		list := a.call(http.MethodGet, "/api/v1/snippets", "", nil).jsonList(t)
		require.Len(t, list, 1)
		assert.Equal(t, long["id"], list[0]["id"])
	})

	t.Run("the list shows only the ten newest, newest first", func(t *testing.T) {
		b := newAPI(t)
		tok := b.issueToken(2)
		var ids []float64
		for i := range 12 {
			ids = append(ids, b.createSnippet(tok, fmt.Sprintf("snippet %d", i), 7).json(t)["id"].(float64))
		}

		list := b.call(http.MethodGet, "/api/v1/snippets", "", nil).jsonList(t)

		require.Len(t, list, 10)
		assert.Equal(t, ids[11], list[0]["id"])
		assert.Equal(t, ids[2], list[9]["id"])
	})
}

func TestE2E_HTTPBehaviour(t *testing.T) {
	a := newAPI(t)

	t.Run("every response is tagged with a request id", func(t *testing.T) {
		generated := a.call(http.MethodGet, "/ping", "", nil)
		assert.Len(t, generated.headers.Get("X-Request-ID"), 16)

		echoed := a.send(http.MethodGet, "/ping", map[string]string{"X-Request-ID": "trace-42"}, nil)
		assert.Equal(t, "trace-42", echoed.headers.Get("X-Request-ID"))

		notFound := a.call(http.MethodGet, "/nope", "", nil)
		assert.NotEmpty(t, notFound.headers.Get("X-Request-ID"))
	})

	t.Run("security headers are always present", func(t *testing.T) {
		resp := a.call(http.MethodGet, "/api/v1/snippets", "", nil)

		assert.Equal(t, "nosniff", resp.headers.Get("X-Content-Type-Options"))
		assert.Equal(t, "DENY", resp.headers.Get("X-Frame-Options"))
		assert.Equal(t, "no-referrer", resp.headers.Get("Referrer-Policy"))
		assert.Contains(t, resp.headers.Get("Content-Security-Policy"), "default-src 'none'")
	})

	t.Run("CORS preflight from the allowed site", func(t *testing.T) {
		resp := a.send(http.MethodOptions, "/api/v1/snippets", map[string]string{
			"Origin":                         allowedSite,
			"Access-Control-Request-Method":  "POST",
			"Access-Control-Request-Headers": "authorization,content-type",
		}, nil)

		assert.Equal(t, http.StatusNoContent, resp.status)
		assert.Equal(t, allowedSite, resp.headers.Get("Access-Control-Allow-Origin"))
		assert.Contains(t, resp.headers.Get("Access-Control-Allow-Methods"), "POST")
		assert.Contains(t, resp.headers.Get("Access-Control-Allow-Headers"), "Authorization")
	})

	t.Run("CORS denies other sites but still serves them", func(t *testing.T) {
		preflight := a.send(http.MethodOptions, "/api/v1/snippets", map[string]string{
			"Origin": "https://evil.test", "Access-Control-Request-Method": "POST",
		}, nil)
		assert.Empty(t, preflight.headers.Get("Access-Control-Allow-Origin"))

		simple := a.send(http.MethodGet, "/ping", map[string]string{"Origin": "https://evil.test"}, nil)
		assert.Equal(t, http.StatusOK, simple.status)
		assert.Empty(t, simple.headers.Get("Access-Control-Allow-Origin"))
	})

	t.Run("CORS exposes Location to the allowed site", func(t *testing.T) {
		resp := a.send(http.MethodGet, "/ping", map[string]string{"Origin": allowedSite}, nil)

		assert.Equal(t, allowedSite, resp.headers.Get("Access-Control-Allow-Origin"))
		assert.Contains(t, resp.headers.Get("Access-Control-Expose-Headers"), "Location")
	})

	t.Run("unknown routes and wrong methods use the JSON error envelope", func(t *testing.T) {
		missing := a.call(http.MethodGet, "/nope", "", nil)
		assert.Equal(t, http.StatusNotFound, missing.status)
		assert.Equal(t, "NOT_FOUND", missing.json(t)["error"].(map[string]any)["code"])

		wrong := a.call(http.MethodDelete, "/api/v1/snippets", "", nil)
		assert.Equal(t, http.StatusMethodNotAllowed, wrong.status)
		assert.Equal(t, "METHOD_NOT_ALLOWED", wrong.json(t)["error"].(map[string]any)["code"])
	})
}

func TestE2E_OpenAPIDescribesTheLiveRoutes(t *testing.T) {
	a := newAPI(t)

	spec := a.call(http.MethodGet, "/swagger/openapi.json", "", nil).json(t)
	paths, ok := spec["paths"].(map[string]any)
	require.True(t, ok)
	require.NotEmpty(t, paths)

	for path, item := range paths {
		methods, ok := item.(map[string]any)
		require.True(t, ok, path)

		concrete := strings.NewReplacer("{id}", "1").Replace(path)
		for method := range methods {
			upper := strings.ToUpper(method)
			if !strings.Contains("GET POST PUT PATCH DELETE", upper) {
				continue
			}

			resp := a.call(upper, concrete, "", map[string]any{})

			assert.NotEqual(t, http.StatusMethodNotAllowed, resp.status, "%s %s is documented but not routed", upper, path)
			if !strings.Contains(path, "{") {
				assert.NotEqual(t, http.StatusNotFound, resp.status, "%s %s is documented but not routed", upper, path)
			}
		}
	}
}

func TestE2E_ConcurrentUse(t *testing.T) {
	a := newAPI(t)
	token := a.issueToken(1)

	const writers = 20
	statuses := make([]int, writers)
	ids := make([]float64, writers)

	var wg sync.WaitGroup
	for i := range writers {
		wg.Add(1)
		go func() {
			defer wg.Done()
			resp := a.createSnippet(token, fmt.Sprintf("snippet %d", i), 7)
			statuses[i] = resp.status
			ids[i], _ = resp.json(t)["id"].(float64)
		}()
	}
	wg.Wait()

	seen := map[float64]bool{}
	for i := range writers {
		assert.Equal(t, http.StatusCreated, statuses[i])
		seen[ids[i]] = true
	}
	assert.Len(t, seen, writers, "every concurrent create must get its own id")
}

func TestE2E_ProductionDefaults(t *testing.T) {
	cfg, err := config.Load(func(key string) string {
		if key == "JWT_SECRET" {
			return jwtSecret
		}
		return ""
	})
	require.NoError(t, err)

	server := httptest.NewServer(app.NewHandler(app.Dependencies{Config: cfg, DB: testutil.NewSQLiteDB(t)}))
	t.Cleanup(server.Close)

	token := testutil.IssueToken(t, jwtSecret, 1, time.Now().UTC(), time.Hour)
	req, err := http.NewRequest(http.MethodPost, server.URL+"/api/v1/snippets", strings.NewReader(`{"title":"Hello","content":"World","expires":1}`))
	require.NoError(t, err)
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	t.Cleanup(func() { _ = resp.Body.Close() })
	require.Equal(t, http.StatusCreated, resp.StatusCode)

	var snippet struct {
		ID      uint      `json:"id"`
		Created time.Time `json:"created"`
		Expires time.Time `json:"expires"`
	}
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&snippet))
	assert.NotZero(t, snippet.ID)
	assert.WithinDuration(t, time.Now().UTC(), snippet.Created, time.Minute)
	assert.Equal(t, 24*time.Hour, snippet.Expires.Sub(snippet.Created))
}
