//go:build e2e

package e2e_test

import (
	"bytes"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"

	"snippetbox/backend-user/internal/app"
	"snippetbox/backend-user/internal/config"
	"snippetbox/backend-user/internal/security"
	"snippetbox/backend-user/internal/testutil"
)

const (
	jwtSecret   = "0123456789abcdef0123456789abcdef"
	allowedSite = "https://app.test"
)

type api struct {
	t      *testing.T
	server *httptest.Server
	client *http.Client
	clock  *testutil.FixedClock
}

type response struct {
	status  int
	headers http.Header
	body    []byte
}

func (r response) json(t *testing.T) map[string]any {
	t.Helper()
	var out map[string]any
	require.NoError(t, json.Unmarshal(r.body, &out), "body: %s", r.body)
	return out
}

func newAPI(t *testing.T) *api {
	t.Helper()

	env := map[string]string{
		"JWT_SECRET":           jwtSecret,
		"JWT_TTL":              "12h",
		"CORS_ALLOWED_ORIGINS": allowedSite,
	}
	cfg, err := config.Load(func(key string) string { return env[key] })
	require.NoError(t, err)

	clock := testutil.NewFixedClock(testutil.Epoch)
	handler := app.NewHandler(app.Dependencies{
		Config: cfg,
		DB:     testutil.NewSQLiteDB(t),
		Logger: slog.New(slog.NewTextHandler(io.Discard, nil)),
		Clock:  clock,
		Hasher: security.NewBcryptHasher(bcrypt.MinCost),
	})

	server := httptest.NewServer(handler)
	t.Cleanup(server.Close)

	return &api{
		t:      t,
		server: server,
		clock:  clock,
		client: &http.Client{
			Timeout: 10 * time.Second,
			CheckRedirect: func(*http.Request, []*http.Request) error {
				return http.ErrUseLastResponse
			},
		},
	}
}

func (a *api) send(method, path string, headers map[string]string, body io.Reader) response {
	a.t.Helper()

	req, err := http.NewRequest(method, a.server.URL+path, body)
	require.NoError(a.t, err)
	for key, value := range headers {
		req.Header.Set(key, value)
	}

	resp, err := a.client.Do(req)
	require.NoError(a.t, err)
	defer resp.Body.Close()

	raw, err := io.ReadAll(resp.Body)
	require.NoError(a.t, err)
	return response{status: resp.StatusCode, headers: resp.Header, body: raw}
}

func (a *api) call(method, path, token string, payload any) response {
	a.t.Helper()

	var body io.Reader
	headers := map[string]string{}
	if payload != nil {
		raw, err := json.Marshal(payload)
		require.NoError(a.t, err)
		body = bytes.NewReader(raw)
		headers["Content-Type"] = "application/json"
	}
	if token != "" {
		headers["Authorization"] = "Bearer " + token
	}
	return a.send(method, path, headers, body)
}

func (a *api) raw(method, path, token, body string) response {
	a.t.Helper()

	headers := map[string]string{"Content-Type": "application/json"}
	if token != "" {
		headers["Authorization"] = "Bearer " + token
	}
	return a.send(method, path, headers, bytes.NewBufferString(body))
}

func (a *api) signup(name, email, password string) response {
	a.t.Helper()
	return a.call(http.MethodPost, "/api/v1/auth/signup", "", map[string]string{"name": name, "email": email, "password": password})
}

func (a *api) login(email, password string) response {
	a.t.Helper()
	return a.call(http.MethodPost, "/api/v1/auth/login", "", map[string]string{"email": email, "password": password})
}

func (a *api) registerAndLogin(email string) string {
	a.t.Helper()

	require.Equal(a.t, http.StatusCreated, a.signup("Ada Lovelace", email, "pa55word").status)
	resp := a.login(email, "pa55word")
	require.Equal(a.t, http.StatusOK, resp.status)

	token, ok := resp.json(a.t)["token"].(string)
	require.True(a.t, ok)
	return token
}
