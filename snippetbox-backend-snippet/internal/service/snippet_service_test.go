package service_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"snippetbox/backend-snippet/internal/apperror"
	"snippetbox/backend-snippet/internal/dto"
	"snippetbox/backend-snippet/internal/model"
	"snippetbox/backend-snippet/internal/service"
	"snippetbox/backend-snippet/internal/testutil"
)

type snippetFixture struct {
	svc   *service.SnippetService
	repo  *snippetRepoStub
	clock *testutil.FixedClock
}

func newSnippetFixture() *snippetFixture {
	f := &snippetFixture{
		repo:  &snippetRepoStub{MemSnippetRepo: testutil.NewMemSnippetRepo()},
		clock: testutil.NewFixedClock(testutil.Epoch),
	}
	f.svc = service.NewSnippetService(f.repo, f.clock)
	return f
}

var (
	author     = &model.Principal{ID: 1}
	snippetReq = dto.CreateSnippetRequest{Title: "Hello", Content: "World", Expires: 7}
)

func TestSnippetService_Create(t *testing.T) {
	t.Parallel()

	t.Run("stores a snippet that expires after the requested days", func(t *testing.T) {
		t.Parallel()
		f := newSnippetFixture()

		snippet, err := f.svc.Create(context.Background(), author, snippetReq)

		require.NoError(t, err)
		assert.NotZero(t, snippet.ID)
		assert.Equal(t, testutil.Epoch, snippet.Created)
		assert.Equal(t, testutil.Epoch.Add(7*24*time.Hour), snippet.Expires)
		assert.Equal(t, 1, f.repo.Count())
	})

	t.Run("requires an authenticated actor", func(t *testing.T) {
		t.Parallel()
		f := newSnippetFixture()

		_, err := f.svc.Create(context.Background(), nil, snippetReq)

		assert.ErrorIs(t, err, apperror.ErrUnauthenticated)
		assert.Zero(t, f.repo.Count())
	})

	t.Run("rejects invalid input", func(t *testing.T) {
		t.Parallel()
		f := newSnippetFixture()

		_, err := f.svc.Create(context.Background(), author, dto.CreateSnippetRequest{})

		var verr *apperror.ValidationError
		require.ErrorAs(t, err, &verr)
		assert.Contains(t, verr.FieldErrors, "title")
		assert.Zero(t, f.repo.Count())
	})

	t.Run("wraps repository failures", func(t *testing.T) {
		t.Parallel()
		f := newSnippetFixture()
		f.repo.err = errBoom

		_, err := f.svc.Create(context.Background(), author, snippetReq)

		require.ErrorIs(t, err, errBoom)
		assert.Contains(t, err.Error(), "create snippet")
	})
}

func TestSnippetService_Get(t *testing.T) {
	t.Parallel()

	t.Run("returns an active snippet", func(t *testing.T) {
		t.Parallel()
		f := newSnippetFixture()
		created, err := f.svc.Create(context.Background(), author, snippetReq)
		require.NoError(t, err)

		got, err := f.svc.Get(context.Background(), created.ID)

		require.NoError(t, err)
		assert.Equal(t, "Hello", got.Title)
	})

	t.Run("treats id zero, unknown ids and expired snippets as not found", func(t *testing.T) {
		t.Parallel()
		f := newSnippetFixture()
		created, err := f.svc.Create(context.Background(), author, dto.CreateSnippetRequest{Title: "t", Content: "c", Expires: 1})
		require.NoError(t, err)

		_, err = f.svc.Get(context.Background(), 0)
		assert.ErrorIs(t, err, apperror.ErrNotFound)

		_, err = f.svc.Get(context.Background(), 999)
		assert.ErrorIs(t, err, apperror.ErrNotFound)

		f.clock.Advance(24 * time.Hour)
		_, err = f.svc.Get(context.Background(), created.ID)
		assert.ErrorIs(t, err, apperror.ErrNotFound)
	})

	t.Run("propagates repository failures", func(t *testing.T) {
		t.Parallel()
		f := newSnippetFixture()
		f.repo.err = errBoom

		_, err := f.svc.Get(context.Background(), 1)

		assert.ErrorIs(t, err, errBoom)
	})
}

func TestSnippetService_ListLatest(t *testing.T) {
	t.Parallel()

	t.Run("is newest first and capped at the configured limit", func(t *testing.T) {
		t.Parallel()
		f := newSnippetFixture()
		for range model.LatestSnippetsLimit + 3 {
			_, err := f.svc.Create(context.Background(), author, snippetReq)
			require.NoError(t, err)
		}

		list, err := f.svc.ListLatest(context.Background())

		require.NoError(t, err)
		require.Len(t, list, model.LatestSnippetsLimit)
		assert.Greater(t, list[0].ID, list[1].ID)
	})

	t.Run("omits expired snippets", func(t *testing.T) {
		t.Parallel()
		f := newSnippetFixture()
		_, err := f.svc.Create(context.Background(), author, dto.CreateSnippetRequest{Title: "short", Content: "c", Expires: 1})
		require.NoError(t, err)
		_, err = f.svc.Create(context.Background(), author, dto.CreateSnippetRequest{Title: "long", Content: "c", Expires: 365})
		require.NoError(t, err)

		f.clock.Advance(48 * time.Hour)
		list, err := f.svc.ListLatest(context.Background())

		require.NoError(t, err)
		require.Len(t, list, 1)
		assert.Equal(t, "long", list[0].Title)
	})

	t.Run("propagates repository failures", func(t *testing.T) {
		t.Parallel()
		f := newSnippetFixture()
		f.repo.err = errBoom

		_, err := f.svc.ListLatest(context.Background())

		assert.ErrorIs(t, err, errBoom)
	})
}
