package server_test

import (
	"context"
	"io"
	"log/slog"
	"net"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"snippetbox/backend-snippet/internal/server"
)

var quiet = slog.New(slog.NewTextHandler(io.Discard, nil))

func listen(t *testing.T) net.Listener {
	t.Helper()
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	return ln
}

func TestNew_SetsTimeouts(t *testing.T) {
	t.Parallel()

	srv := server.New(http.NotFoundHandler())

	assert.NotZero(t, srv.ReadHeaderTimeout)
	assert.NotZero(t, srv.ReadTimeout)
	assert.NotZero(t, srv.WriteTimeout)
	assert.NotZero(t, srv.IdleTimeout)
	assert.NotNil(t, srv.Handler)
}

func TestServe_ServesRequestsAndStopsOnCancel(t *testing.T) {
	t.Parallel()

	ln := listen(t)
	srv := server.New(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = io.WriteString(w, "pong")
	}))

	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error, 1)
	go func() { done <- server.Serve(ctx, srv, ln, quiet) }()

	resp, err := http.Get("http://" + ln.Addr().String())
	require.NoError(t, err)
	body, _ := io.ReadAll(resp.Body)
	_ = resp.Body.Close()
	assert.Equal(t, "pong", string(body))

	cancel()

	select {
	case err := <-done:
		assert.NoError(t, err)
	case <-time.After(5 * time.Second):
		t.Fatal("Serve did not return after the context was cancelled")
	}
}

func TestServe_ShutdownWaitsForInFlightRequests(t *testing.T) {
	t.Parallel()

	ln := listen(t)
	started := make(chan struct{})
	release := make(chan struct{})
	srv := server.New(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		close(started)
		<-release
		_, _ = io.WriteString(w, "finished")
	}))

	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error, 1)
	go func() { done <- server.Serve(ctx, srv, ln, quiet) }()

	type result struct {
		body string
		err  error
	}
	got := make(chan result, 1)
	go func() {
		resp, err := http.Get("http://" + ln.Addr().String())
		if err != nil {
			got <- result{err: err}
			return
		}
		defer resp.Body.Close()
		body, err := io.ReadAll(resp.Body)
		got <- result{body: string(body), err: err}
	}()

	<-started
	cancel()

	select {
	case <-done:
		t.Fatal("Serve returned while a request was still in flight")
	case <-time.After(100 * time.Millisecond):
	}

	close(release)

	res := <-got
	require.NoError(t, res.err)
	assert.Equal(t, "finished", res.body)
	assert.NoError(t, <-done)
}

func TestServe_ReturnsListenerFailures(t *testing.T) {
	t.Parallel()

	ln := listen(t)
	require.NoError(t, ln.Close())

	err := server.Serve(context.Background(), server.New(http.NotFoundHandler()), ln, quiet)

	assert.Error(t, err)
}
