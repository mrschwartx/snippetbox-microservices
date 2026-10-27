//go:build e2e

package e2e_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"snippetbox/backend-user/internal/app"
	"snippetbox/backend-user/internal/config"
	"snippetbox/backend-user/internal/testutil"
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

func TestE2E_UserJourney(t *testing.T) {
	a := newAPI(t)

	signup := a.signup("Ada Lovelace", "Ada@Example.com", "pa55word")
	require.Equal(t, http.StatusCreated, signup.status, string(signup.body))
	user := signup.json(t)
	assert.Equal(t, "ada@example.com", user["email"])
	assert.Equal(t, "Ada Lovelace", user["name"])
	assert.NotContains(t, string(signup.body), "pa55word")
	assert.NotContains(t, string(signup.body), "password")
	assert.NotContains(t, string(signup.body), "$2a$")

	duplicate := a.signup("Someone", "ADA@example.com", "pa55word")
	assert.Equal(t, http.StatusUnprocessableEntity, duplicate.status)
	assert.Equal(t, "Email address is already in use", duplicate.json(t)["error"].(map[string]any)["fields"].(map[string]any)["email"])

	wrong := a.login("ada@example.com", "not-the-password")
	assert.Equal(t, http.StatusUnauthorized, wrong.status)
	assert.Equal(t, "INVALID_CREDENTIALS", wrong.json(t)["error"].(map[string]any)["code"])

	unknown := a.login("nobody@example.com", "pa55word")
	assert.Equal(t, http.StatusUnauthorized, unknown.status)
	assert.JSONEq(t, string(wrong.body), string(unknown.body))

	login := a.login("ADA@example.com", "pa55word")
	require.Equal(t, http.StatusOK, login.status, string(login.body))
	session := login.json(t)
	token := session["token"].(string)
	assert.Equal(t, "Bearer", session["token_type"])
	assert.Equal(t, "2026-10-15T20:00:00Z", session["expires_at"])
	assert.Equal(t, user["id"], session["user"].(map[string]any)["id"])

	me := a.call(http.MethodGet, "/api/v1/auth/me", token, nil)
	assert.Equal(t, http.StatusOK, me.status)
	assert.Equal(t, "ada@example.com", me.json(t)["email"])
	assert.Equal(t, "no-store", me.headers.Get("Cache-Control"))

	logout := a.call(http.MethodPost, "/api/v1/auth/logout", token, nil)
	assert.Equal(t, http.StatusNoContent, logout.status)
	assert.Empty(t, logout.body)
}

func TestE2E_Authentication(t *testing.T) {
	a := newAPI(t)
	token := a.registerAndLogin("ada@example.com")

	protected := []struct{ method, path string }{
		{http.MethodGet, "/api/v1/auth/me"},
		{http.MethodPost, "/api/v1/auth/logout"},
	}

	for _, route := range protected {
		for name, bad := range map[string]string{"missing": "", "garbage": "garbage", "tampered": token + "x"} {
			resp := a.call(route.method, route.path, bad, map[string]any{})

			assert.Equal(t, http.StatusUnauthorized, resp.status, "%s %s (%s)", route.method, route.path, name)
			assert.Equal(t, "UNAUTHENTICATED", resp.json(t)["error"].(map[string]any)["code"])
		}
	}

	notBearer := a.send(http.MethodGet, "/api/v1/auth/me", map[string]string{"Authorization": "Basic " + token}, nil)
	assert.Equal(t, http.StatusUnauthorized, notBearer.status)

	assert.Equal(t, http.StatusOK, a.call(http.MethodGet, "/api/v1/auth/me", token, nil).status)

	a.clock.Advance(11*time.Hour + 59*time.Minute)
	assert.Equal(t, http.StatusOK, a.call(http.MethodGet, "/api/v1/auth/me", token, nil).status)

	a.clock.Advance(2 * time.Minute)
	expired := a.call(http.MethodGet, "/api/v1/auth/me", token, nil)
	assert.Equal(t, http.StatusUnauthorized, expired.status)
}

func TestE2E_RequestValidation(t *testing.T) {
	a := newAPI(t)

	t.Run("signup field errors", func(t *testing.T) {
		tests := []struct {
			name, body, field string
		}{
			{"blank name", `{"name":" ","email":"a@b.co","password":"pa55word"}`, "name"},
			{"bad email", `{"name":"A","email":"nope","password":"pa55word"}`, "email"},
			{"short password", `{"name":"A","email":"a@b.co","password":"short"}`, "password"},
			{"long password", `{"name":"A","email":"a@b.co","password":"` + strings.Repeat("a", 73) + `"}`, "password"},
		}
		for _, tt := range tests {
			resp := a.raw(http.MethodPost, "/api/v1/auth/signup", "", tt.body)

			assert.Equal(t, http.StatusUnprocessableEntity, resp.status, tt.name)
			fields := resp.json(t)["error"].(map[string]any)["fields"].(map[string]any)
			assert.Contains(t, fields, tt.field, tt.name)
		}
	})

	t.Run("login field errors", func(t *testing.T) {
		resp := a.raw(http.MethodPost, "/api/v1/auth/login", "", `{"email":"","password":""}`)

		assert.Equal(t, http.StatusUnprocessableEntity, resp.status)
		fields := resp.json(t)["error"].(map[string]any)["fields"].(map[string]any)
		assert.Contains(t, fields, "email")
		assert.Contains(t, fields, "password")
	})

	t.Run("malformed bodies", func(t *testing.T) {
		for name, body := range map[string]string{
			"not json":       `hello`,
			"empty":          ``,
			"unknown field":  `{"name":"A","email":"a@b.co","password":"pa55word","admin":true}`,
			"wrong type":     `{"name":"A","email":"a@b.co","password":7}`,
			"trailing value": `{"name":"A","email":"a@b.co","password":"pa55word"}{}`,
		} {
			resp := a.raw(http.MethodPost, "/api/v1/auth/signup", "", body)

			assert.Equal(t, http.StatusBadRequest, resp.status, name)
			assert.Equal(t, "BAD_REQUEST", resp.json(t)["error"].(map[string]any)["code"], name)
		}
	})

	t.Run("bodies over 1 MiB", func(t *testing.T) {
		huge := `{"name":"` + strings.Repeat("a", 1<<20) + `","email":"a@b.co","password":"pa55word"}`

		resp := a.raw(http.MethodPost, "/api/v1/auth/signup", "", huge)

		assert.Equal(t, http.StatusRequestEntityTooLarge, resp.status)
		assert.Equal(t, "PAYLOAD_TOO_LARGE", resp.json(t)["error"].(map[string]any)["code"])
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
		resp := a.call(http.MethodPost, "/api/v1/auth/login", "", map[string]string{})

		assert.Equal(t, "nosniff", resp.headers.Get("X-Content-Type-Options"))
		assert.Equal(t, "DENY", resp.headers.Get("X-Frame-Options"))
		assert.Equal(t, "no-referrer", resp.headers.Get("Referrer-Policy"))
		assert.Contains(t, resp.headers.Get("Content-Security-Policy"), "default-src 'none'")
	})

	t.Run("CORS preflight from the allowed site", func(t *testing.T) {
		resp := a.send(http.MethodOptions, "/api/v1/auth/login", map[string]string{
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
		preflight := a.send(http.MethodOptions, "/api/v1/auth/login", map[string]string{
			"Origin": "https://evil.test", "Access-Control-Request-Method": "POST",
		}, nil)
		assert.Empty(t, preflight.headers.Get("Access-Control-Allow-Origin"))

		simple := a.send(http.MethodGet, "/ping", map[string]string{"Origin": "https://evil.test"}, nil)
		assert.Equal(t, http.StatusOK, simple.status)
		assert.Empty(t, simple.headers.Get("Access-Control-Allow-Origin"))
	})

	t.Run("unknown routes and wrong methods use the JSON error envelope", func(t *testing.T) {
		missing := a.call(http.MethodGet, "/nope", "", nil)
		assert.Equal(t, http.StatusNotFound, missing.status)
		assert.Equal(t, "NOT_FOUND", missing.json(t)["error"].(map[string]any)["code"])

		snippets := a.call(http.MethodGet, "/api/v1/snippets", "", nil)
		assert.Equal(t, http.StatusNotFound, snippets.status)

		wrong := a.call(http.MethodGet, "/api/v1/auth/login", "", nil)
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

		for method := range methods {
			upper := strings.ToUpper(method)
			if !strings.Contains("GET POST PUT PATCH DELETE", upper) {
				continue
			}

			resp := a.call(upper, path, "", map[string]any{})

			assert.NotEqual(t, http.StatusMethodNotAllowed, resp.status, "%s %s is documented but not routed", upper, path)
			assert.NotEqual(t, http.StatusNotFound, resp.status, "%s %s is documented but not routed", upper, path)
		}
	}
}

func TestE2E_ConcurrentSignups(t *testing.T) {
	a := newAPI(t)

	const writers = 8
	statuses := make([]int, writers)

	var wg sync.WaitGroup
	for i := range writers {
		wg.Add(1)
		go func() {
			defer wg.Done()
			statuses[i] = a.signup("Ada", "ada@example.com", "pa55word").status
		}()
	}
	wg.Wait()

	created, rejected := 0, 0
	for _, status := range statuses {
		switch status {
		case http.StatusCreated:
			created++
		case http.StatusUnprocessableEntity:
			rejected++
		}
	}
	assert.Equal(t, 1, created)
	assert.Equal(t, writers-1, rejected)
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

	post := func(path, body string) *http.Response {
		resp, err := http.Post(server.URL+path, "application/json", strings.NewReader(body))
		require.NoError(t, err)
		t.Cleanup(func() { _ = resp.Body.Close() })
		return resp
	}

	before := time.Now().UTC()
	require.Equal(t, http.StatusCreated, post("/api/v1/auth/signup", `{"name":"Ada","email":"ada@example.com","password":"pa55word"}`).StatusCode)

	login := post("/api/v1/auth/login", `{"email":"ada@example.com","password":"pa55word"}`)
	require.Equal(t, http.StatusOK, login.StatusCode)

	var session struct {
		Token     string    `json:"token"`
		ExpiresAt time.Time `json:"expires_at"`
	}
	require.NoError(t, json.NewDecoder(login.Body).Decode(&session))
	assert.WithinDuration(t, before.Add(12*time.Hour), session.ExpiresAt, time.Minute)

	req, err := http.NewRequest(http.MethodGet, server.URL+"/api/v1/auth/me", nil)
	require.NoError(t, err)
	req.Header.Set("Authorization", "Bearer "+session.Token)
	me, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	t.Cleanup(func() { _ = me.Body.Close() })
	assert.Equal(t, http.StatusOK, me.StatusCode)
}
