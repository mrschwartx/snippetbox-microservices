package middleware_test

import (
	"bytes"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"snippetbox/backend-snippet/internal/middleware"
)

var ok = http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusNoContent)
})

func serve(h http.Handler, req *http.Request) *httptest.ResponseRecorder {
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	return rec
}

func newLogger() (*slog.Logger, *bytes.Buffer) {
	var buf bytes.Buffer
	return slog.New(slog.NewJSONHandler(&buf, nil)), &buf
}

func logEntries(t *testing.T, buf *bytes.Buffer) []map[string]any {
	t.Helper()
	var entries []map[string]any
	for _, line := range strings.Split(strings.TrimSpace(buf.String()), "\n") {
		if line == "" {
			continue
		}
		var entry map[string]any
		require.NoError(t, json.Unmarshal([]byte(line), &entry))
		entries = append(entries, entry)
	}
	return entries
}

func TestChain_RunsMiddlewareOutermostFirst(t *testing.T) {
	t.Parallel()

	var order []string
	tag := func(name string) middleware.Middleware {
		return func(next http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				order = append(order, name+":in")
				next.ServeHTTP(w, r)
				order = append(order, name+":out")
			})
		}
	}
	final := http.HandlerFunc(func(http.ResponseWriter, *http.Request) { order = append(order, "handler") })

	serve(middleware.Chain(final, tag("a"), tag("b")), httptest.NewRequest(http.MethodGet, "/", nil))

	assert.Equal(t, []string{"a:in", "b:in", "handler", "b:out", "a:out"}, order)
}

func TestChain_WithoutMiddlewareReturnsHandler(t *testing.T) {
	t.Parallel()

	rec := serve(middleware.Chain(ok), httptest.NewRequest(http.MethodGet, "/", nil))

	assert.Equal(t, http.StatusNoContent, rec.Code)
}

func TestRequestID(t *testing.T) {
	t.Parallel()

	var seen string
	capture := http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) {
		seen = r.Header.Get(middleware.RequestIDHeader)
	})
	h := middleware.RequestID()(capture)

	t.Run("generates an id when absent", func(t *testing.T) {
		rec := serve(h, httptest.NewRequest(http.MethodGet, "/", nil))

		assert.Len(t, rec.Header().Get(middleware.RequestIDHeader), 16)
		assert.Equal(t, rec.Header().Get(middleware.RequestIDHeader), seen)
	})

	t.Run("keeps a client supplied id", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.Header.Set(middleware.RequestIDHeader, "client-id-1")

		rec := serve(h, req)

		assert.Equal(t, "client-id-1", rec.Header().Get(middleware.RequestIDHeader))
		assert.Equal(t, "client-id-1", seen)
	})

	t.Run("replaces an oversized id", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.Header.Set(middleware.RequestIDHeader, strings.Repeat("x", 65))

		rec := serve(h, req)

		assert.Len(t, rec.Header().Get(middleware.RequestIDHeader), 16)
	})

	t.Run("generates distinct ids", func(t *testing.T) {
		first := serve(h, httptest.NewRequest(http.MethodGet, "/", nil)).Header().Get(middleware.RequestIDHeader)
		second := serve(h, httptest.NewRequest(http.MethodGet, "/", nil)).Header().Get(middleware.RequestIDHeader)

		assert.NotEqual(t, first, second)
	})
}

func TestAccessLog(t *testing.T) {
	t.Parallel()

	t.Run("records status, size and request id", func(t *testing.T) {
		log, buf := newLogger()
		h := middleware.Chain(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusCreated)
			_, _ = w.Write([]byte("hello"))
		}), middleware.RequestID(), middleware.AccessLog(log))

		req := httptest.NewRequest(http.MethodPost, "/things", nil)
		req.Header.Set(middleware.RequestIDHeader, "abc")
		serve(h, req)

		entries := logEntries(t, buf)
		require.Len(t, entries, 1)
		assert.Equal(t, "request", entries[0]["msg"])
		assert.Equal(t, "POST", entries[0]["method"])
		assert.Equal(t, "/things", entries[0]["path"])
		assert.EqualValues(t, http.StatusCreated, entries[0]["status"])
		assert.EqualValues(t, 5, entries[0]["bytes"])
		assert.Equal(t, "abc", entries[0]["request_id"])
	})

	t.Run("defaults to 200 when the handler writes a body only", func(t *testing.T) {
		log, buf := newLogger()
		h := middleware.AccessLog(log)(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			_, _ = w.Write([]byte("x"))
		}))

		serve(h, httptest.NewRequest(http.MethodGet, "/", nil))

		assert.EqualValues(t, http.StatusOK, logEntries(t, buf)[0]["status"])
	})

	t.Run("defaults to 200 when the handler writes nothing", func(t *testing.T) {
		log, buf := newLogger()

		serve(middleware.AccessLog(log)(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {})),
			httptest.NewRequest(http.MethodGet, "/", nil))

		assert.EqualValues(t, http.StatusOK, logEntries(t, buf)[0]["status"])
	})

	t.Run("keeps the first status when WriteHeader is called twice", func(t *testing.T) {
		log, buf := newLogger()
		h := middleware.AccessLog(log)(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusAccepted)
			w.WriteHeader(http.StatusTeapot)
		}))

		serve(h, httptest.NewRequest(http.MethodGet, "/", nil))

		assert.EqualValues(t, http.StatusAccepted, logEntries(t, buf)[0]["status"])
	})

	t.Run("exposes the underlying writer for http.ResponseController", func(t *testing.T) {
		log, _ := newLogger()
		var flushErr error
		h := middleware.AccessLog(log)(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			flushErr = http.NewResponseController(w).Flush()
		}))

		serve(h, httptest.NewRequest(http.MethodGet, "/", nil))

		assert.NoError(t, flushErr)
	})
}

func TestRecoverer(t *testing.T) {
	t.Parallel()

	t.Run("passes requests through when nothing panics", func(t *testing.T) {
		log, buf := newLogger()

		rec := serve(middleware.Recoverer(log)(ok), httptest.NewRequest(http.MethodGet, "/", nil))

		assert.Equal(t, http.StatusNoContent, rec.Code)
		assert.Empty(t, buf.String())
	})

	t.Run("turns a panic into a JSON 500 and logs it", func(t *testing.T) {
		log, buf := newLogger()
		h := middleware.Chain(http.HandlerFunc(func(http.ResponseWriter, *http.Request) { panic("kaboom") }),
			middleware.RequestID(), middleware.Recoverer(log))

		rec := serve(h, httptest.NewRequest(http.MethodGet, "/boom", nil))

		assert.Equal(t, http.StatusInternalServerError, rec.Code)
		assert.Equal(t, "application/json", rec.Header().Get("Content-Type"))
		assert.Equal(t, "close", rec.Header().Get("Connection"))
		assert.JSONEq(t, `{"error":{"code":"INTERNAL_ERROR","message":"Internal server error"}}`, rec.Body.String())
		assert.NotContains(t, rec.Body.String(), "kaboom")

		entries := logEntries(t, buf)
		require.Len(t, entries, 1)
		assert.Equal(t, "kaboom", entries[0]["panic"])
		assert.Equal(t, "/boom", entries[0]["path"])
		assert.NotEmpty(t, entries[0]["request_id"])
	})

	t.Run("lets http.ErrAbortHandler through", func(t *testing.T) {
		log, _ := newLogger()
		h := middleware.Recoverer(log)(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
			panic(http.ErrAbortHandler)
		}))

		assert.PanicsWithValue(t, http.ErrAbortHandler, func() {
			serve(h, httptest.NewRequest(http.MethodGet, "/", nil))
		})
	})
}

func TestSecureHeaders(t *testing.T) {
	t.Parallel()

	h := middleware.SecureHeaders()(ok)

	t.Run("api paths get a locked down policy", func(t *testing.T) {
		rec := serve(h, httptest.NewRequest(http.MethodGet, "/api/v1/snippets", nil))

		assert.Equal(t, "nosniff", rec.Header().Get("X-Content-Type-Options"))
		assert.Equal(t, "DENY", rec.Header().Get("X-Frame-Options"))
		assert.Equal(t, "no-referrer", rec.Header().Get("Referrer-Policy"))
		assert.Equal(t, "default-src 'none'; frame-ancestors 'none'", rec.Header().Get("Content-Security-Policy"))
	})

	t.Run("swagger paths skip the CSP so the UI can load", func(t *testing.T) {
		for _, path := range []string{"/swagger", "/swagger/", "/swagger/openapi.json"} {
			rec := serve(h, httptest.NewRequest(http.MethodGet, path, nil))

			assert.Empty(t, rec.Header().Get("Content-Security-Policy"), path)
			assert.Equal(t, "nosniff", rec.Header().Get("X-Content-Type-Options"), path)
		}
	})
}

func TestCORS(t *testing.T) {
	t.Parallel()

	const allowed = "https://app.test"
	named := middleware.CORS([]string{allowed})(ok)
	wildcard := middleware.CORS([]string{"*"})(ok)

	request := func(method, origin string, preflight bool) *http.Request {
		req := httptest.NewRequest(method, "/api/v1/snippets", nil)
		if origin != "" {
			req.Header.Set("Origin", origin)
		}
		if preflight {
			req.Header.Set("Access-Control-Request-Method", "POST")
		}
		return req
	}

	t.Run("requests without an Origin are untouched", func(t *testing.T) {
		rec := serve(named, request(http.MethodGet, "", false))

		assert.Equal(t, http.StatusNoContent, rec.Code)
		assert.Empty(t, rec.Header().Get("Vary"))
		assert.Empty(t, rec.Header().Get("Access-Control-Allow-Origin"))
	})

	t.Run("an allowed origin is echoed back", func(t *testing.T) {
		rec := serve(named, request(http.MethodGet, allowed, false))

		assert.Equal(t, http.StatusNoContent, rec.Code)
		assert.Equal(t, allowed, rec.Header().Get("Access-Control-Allow-Origin"))
		assert.Equal(t, "Origin", rec.Header().Get("Vary"))
		assert.Equal(t, "X-Request-ID, Location", rec.Header().Get("Access-Control-Expose-Headers"))
	})

	t.Run("an unknown origin gets no CORS headers but is still served", func(t *testing.T) {
		rec := serve(named, request(http.MethodGet, "https://evil.test", false))

		assert.Equal(t, http.StatusNoContent, rec.Code)
		assert.Empty(t, rec.Header().Get("Access-Control-Allow-Origin"))
		assert.Equal(t, "Origin", rec.Header().Get("Vary"))
	})

	t.Run("a wildcard allows any origin", func(t *testing.T) {
		rec := serve(wildcard, request(http.MethodGet, "https://anything.test", false))

		assert.Equal(t, "*", rec.Header().Get("Access-Control-Allow-Origin"))
	})

	t.Run("a preflight from an allowed origin is answered directly", func(t *testing.T) {
		rec := serve(middleware.CORS([]string{allowed})(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
			t.Error("preflight must not reach the handler")
		})), request(http.MethodOptions, allowed, true))

		assert.Equal(t, http.StatusNoContent, rec.Code)
		assert.Equal(t, "GET, POST, OPTIONS", rec.Header().Get("Access-Control-Allow-Methods"))
		assert.Equal(t, "Authorization, Content-Type", rec.Header().Get("Access-Control-Allow-Headers"))
		assert.Equal(t, "600", rec.Header().Get("Access-Control-Max-Age"))
		assert.Equal(t, allowed, rec.Header().Get("Access-Control-Allow-Origin"))
	})

	t.Run("a preflight from a disallowed origin is denied without permissions", func(t *testing.T) {
		rec := serve(named, request(http.MethodOptions, "https://evil.test", true))

		assert.Equal(t, http.StatusNoContent, rec.Code)
		assert.Empty(t, rec.Header().Get("Access-Control-Allow-Origin"))
		assert.Empty(t, rec.Header().Get("Access-Control-Allow-Methods"))
	})

	t.Run("an OPTIONS request that is not a preflight reaches the handler", func(t *testing.T) {
		reached := false
		h := middleware.CORS([]string{allowed})(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			reached = true
			w.WriteHeader(http.StatusTeapot)
		}))

		rec := serve(h, request(http.MethodOptions, allowed, false))

		assert.True(t, reached)
		assert.Equal(t, http.StatusTeapot, rec.Code)
	})
}

func TestBodyLimit(t *testing.T) {
	t.Parallel()

	read := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if _, err := io.ReadAll(r.Body); err != nil {
			w.WriteHeader(http.StatusRequestEntityTooLarge)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	})
	h := middleware.BodyLimit(10)(read)

	t.Run("accepts a body at the limit", func(t *testing.T) {
		rec := serve(h, httptest.NewRequest(http.MethodPost, "/", strings.NewReader(strings.Repeat("a", 10))))
		assert.Equal(t, http.StatusNoContent, rec.Code)
	})

	t.Run("rejects a body over the limit", func(t *testing.T) {
		rec := serve(h, httptest.NewRequest(http.MethodPost, "/", strings.NewReader(strings.Repeat("a", 11))))
		assert.Equal(t, http.StatusRequestEntityTooLarge, rec.Code)
	})

	t.Run("tolerates a nil body", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.Body = nil

		rec := serve(middleware.BodyLimit(10)(ok), req)

		assert.Equal(t, http.StatusNoContent, rec.Code)
	})
}
