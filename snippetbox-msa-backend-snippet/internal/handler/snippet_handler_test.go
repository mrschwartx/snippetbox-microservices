package handler_test

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/labstack/echo/v5"
	"github.com/labstack/echo/v5/echotest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"snippetbox/backend-snippet/internal/apperror"
	"snippetbox/backend-snippet/internal/dto"
	"snippetbox/backend-snippet/internal/handler"
	"snippetbox/backend-snippet/internal/middleware"
	"snippetbox/backend-snippet/internal/model"
)

func sampleSnippet(id uint) model.Snippet {
	return model.Snippet{ID: id, Title: "Hello", Content: "World", Created: created, Expires: created.Add(24 * time.Hour)}
}

func pathID(raw string) echo.PathValues {
	return echo.PathValues{{Name: "id", Value: raw}}
}

func TestSnippetHandler_List(t *testing.T) {
	t.Parallel()

	t.Run("returns snippets in service order", func(t *testing.T) {
		t.Parallel()
		h := handler.NewSnippetHandler(snippetServiceFake{listLatest: func(context.Context) ([]model.Snippet, error) {
			return []model.Snippet{sampleSnippet(2), sampleSnippet(1)}, nil
		}})

		status, body, _ := serveHandler(t, h.List, echotest.ContextConfig{})

		assert.Equal(t, http.StatusOK, status)
		assert.JSONEq(t, `[
			{"id":2,"title":"Hello","content":"World","created":"2026-10-15T08:00:00Z","expires":"2026-10-16T08:00:00Z"},
			{"id":1,"title":"Hello","content":"World","created":"2026-10-15T08:00:00Z","expires":"2026-10-16T08:00:00Z"}
		]`, body)
	})

	t.Run("an empty result is an empty array, never null", func(t *testing.T) {
		t.Parallel()
		h := handler.NewSnippetHandler(snippetServiceFake{listLatest: func(context.Context) ([]model.Snippet, error) {
			return nil, nil
		}})

		_, body, _ := serveHandler(t, h.List, echotest.ContextConfig{})

		assert.JSONEq(t, `[]`, body)
	})

	t.Run("service failures become 500", func(t *testing.T) {
		t.Parallel()
		h := handler.NewSnippetHandler(snippetServiceFake{listLatest: func(context.Context) ([]model.Snippet, error) {
			return nil, assert.AnError
		}})

		status, _, _ := serveHandler(t, h.List, echotest.ContextConfig{})

		assert.Equal(t, http.StatusInternalServerError, status)
	})
}

func TestSnippetHandler_Get(t *testing.T) {
	t.Parallel()

	t.Run("passes the parsed id to the service", func(t *testing.T) {
		t.Parallel()
		var gotID uint
		h := handler.NewSnippetHandler(snippetServiceFake{get: func(_ context.Context, id uint) (*model.Snippet, error) {
			gotID = id
			s := sampleSnippet(id)
			return &s, nil
		}})

		status, body, _ := serveHandler(t, h.Get, echotest.ContextConfig{PathValues: pathID("42")})

		assert.Equal(t, http.StatusOK, status)
		assert.Equal(t, uint(42), gotID)
		assert.Contains(t, body, `"id":42`)
	})

	t.Run("a non numeric or out of range id is not found without calling the service", func(t *testing.T) {
		t.Parallel()
		h := handler.NewSnippetHandler(snippetServiceFake{get: func(context.Context, uint) (*model.Snippet, error) {
			t.Fatal("service must not be called")
			return nil, nil
		}})

		for _, raw := range []string{"abc", "-1", "+1", "1.5", "", " 1", "9223372036854775808", "99999999999999999999999"} {
			status, body, _ := serveHandler(t, h.Get, echotest.ContextConfig{PathValues: pathID(raw)})

			assert.Equal(t, http.StatusNotFound, status, "id %q", raw)
			assert.Contains(t, body, `"NOT_FOUND"`, "id %q", raw)
		}
	})

	t.Run("accepts the largest id the database can store", func(t *testing.T) {
		t.Parallel()
		var gotID uint
		h := handler.NewSnippetHandler(snippetServiceFake{get: func(_ context.Context, id uint) (*model.Snippet, error) {
			gotID = id
			return nil, apperror.ErrNotFound
		}})

		serveHandler(t, h.Get, echotest.ContextConfig{PathValues: pathID("9223372036854775807")})

		assert.Equal(t, uint(1<<63-1), gotID)
	})

	t.Run("service not found becomes 404", func(t *testing.T) {
		t.Parallel()
		h := handler.NewSnippetHandler(snippetServiceFake{get: func(context.Context, uint) (*model.Snippet, error) {
			return nil, apperror.ErrNotFound
		}})

		status, _, _ := serveHandler(t, h.Get, echotest.ContextConfig{PathValues: pathID("1")})

		assert.Equal(t, http.StatusNotFound, status)
	})
}

func TestSnippetHandler_Create(t *testing.T) {
	t.Parallel()

	actor := &model.Principal{ID: 8}

	t.Run("creates the snippet for the authenticated user and sets Location", func(t *testing.T) {
		t.Parallel()
		var gotActor *model.Principal
		var gotReq dto.CreateSnippetRequest
		h := handler.NewSnippetHandler(snippetServiceFake{create: func(_ context.Context, a *model.Principal, req dto.CreateSnippetRequest) (*model.Snippet, error) {
			gotActor, gotReq = a, req
			s := sampleSnippet(11)
			return &s, nil
		}})

		c, rec := echotest.ContextConfig{JSONBody: []byte(`{"title":"Hello","content":"World","expires":7}`)}.ToContextRecorder(t)
		c.Set(middleware.PrincipalContextKey, actor)

		require.NoError(t, h.Create(c))

		assert.Equal(t, http.StatusCreated, rec.Code)
		assert.Equal(t, "/api/v1/snippets/11", rec.Header().Get("Location"))
		assert.Contains(t, rec.Body.String(), `"id":11`)
		assert.Same(t, actor, gotActor)
		assert.Equal(t, dto.CreateSnippetRequest{Title: "Hello", Content: "World", Expires: 7}, gotReq)
	})

	t.Run("rejects unknown fields and malformed bodies", func(t *testing.T) {
		t.Parallel()
		h := handler.NewSnippetHandler(snippetServiceFake{})

		for _, body := range []string{`{"title":"t","extra":1}`, `{`, ``} {
			status, _, _ := serveHandler(t, h.Create, echotest.ContextConfig{JSONBody: []byte(body)})

			assert.Equal(t, http.StatusBadRequest, status, body)
		}
	})

	t.Run("service errors are mapped and no Location is set", func(t *testing.T) {
		t.Parallel()
		h := handler.NewSnippetHandler(snippetServiceFake{create: func(context.Context, *model.Principal, dto.CreateSnippetRequest) (*model.Snippet, error) {
			return nil, apperror.NewFieldValidationError("title", "blank")
		}})

		status, body, headers := serveHandler(t, h.Create, echotest.ContextConfig{JSONBody: []byte(`{}`)})

		assert.Equal(t, http.StatusUnprocessableEntity, status)
		assert.Contains(t, body, `"title":"blank"`)
		assert.Empty(t, headers.Get("Location"))
	})
}
