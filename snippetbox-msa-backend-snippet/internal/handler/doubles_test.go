package handler_test

import (
	"bytes"
	"context"
	"log/slog"
	"net/http"
	"testing"
	"time"

	"github.com/labstack/echo/v5"
	"github.com/labstack/echo/v5/echotest"

	"snippetbox/backend-snippet/internal/dto"
	"snippetbox/backend-snippet/internal/handler"
	"snippetbox/backend-snippet/internal/model"
)

var created = time.Date(2026, 10, 15, 8, 0, 0, 0, time.UTC)

type snippetServiceFake struct {
	create     func(context.Context, *model.Principal, dto.CreateSnippetRequest) (*model.Snippet, error)
	get        func(context.Context, uint) (*model.Snippet, error)
	listLatest func(context.Context) ([]model.Snippet, error)
}

func (f snippetServiceFake) Create(ctx context.Context, actor *model.Principal, req dto.CreateSnippetRequest) (*model.Snippet, error) {
	return f.create(ctx, actor, req)
}

func (f snippetServiceFake) Get(ctx context.Context, id uint) (*model.Snippet, error) {
	return f.get(ctx, id)
}

func (f snippetServiceFake) ListLatest(ctx context.Context) ([]model.Snippet, error) {
	return f.listLatest(ctx)
}

func newLogger() (*slog.Logger, *bytes.Buffer) {
	var buf bytes.Buffer
	return slog.New(slog.NewJSONHandler(&buf, nil)), &buf
}

func serveHandler(t *testing.T, h echo.HandlerFunc, cfg echotest.ContextConfig) (int, string, http.Header) {
	t.Helper()
	log, _ := newLogger()
	rec := cfg.ServeWithHandler(t, h, handler.NewErrorHandler(log))
	return rec.Code, rec.Body.String(), rec.Header()
}
