//go:build integration

package integration_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"snippetbox/backend-snippet/internal/apperror"
	"snippetbox/backend-snippet/internal/database"
	"snippetbox/backend-snippet/internal/model"
	"snippetbox/backend-snippet/internal/repository"
	"snippetbox/backend-snippet/internal/testutil"
)

func newDB(t *testing.T) *gorm.DB {
	t.Helper()
	return testutil.NewSQLiteDB(t)
}

func newSnippet(title string, days int) *model.Snippet {
	s := model.NewSnippet(title, "content of "+title, days, testutil.Epoch)
	return &s
}

func TestSnippetRepository_CreateAndFind(t *testing.T) {
	ctx := context.Background()

	t.Run("round trips a snippet in UTC", func(t *testing.T) {
		repo := repository.NewSnippetRepository(newDB(t))
		snippet := newSnippet("hello", 7)

		require.NoError(t, repo.Create(ctx, snippet))
		found, err := repo.FindActiveByID(ctx, snippet.ID, testutil.Epoch)

		require.NoError(t, err)
		assert.Equal(t, "hello", found.Title)
		assert.Equal(t, "content of hello", found.Content)
		assert.True(t, snippet.Created.Equal(found.Created))
		assert.True(t, snippet.Expires.Equal(found.Expires))
		assert.Equal(t, time.UTC, found.Created.Location())
		assert.Equal(t, time.UTC, found.Expires.Location())
	})

	t.Run("stores content beyond the varchar limit", func(t *testing.T) {
		repo := repository.NewSnippetRepository(newDB(t))
		snippet := newSnippet("big", 1)
		snippet.Content = stringOfLen(model.ContentMaxBytes)

		require.NoError(t, repo.Create(ctx, snippet))
		found, err := repo.FindActiveByID(ctx, snippet.ID, testutil.Epoch)

		require.NoError(t, err)
		assert.Len(t, found.Content, model.ContentMaxBytes)
	})

	t.Run("a snippet stops being visible exactly when it expires", func(t *testing.T) {
		repo := repository.NewSnippetRepository(newDB(t))
		snippet := newSnippet("short", 1)
		require.NoError(t, repo.Create(ctx, snippet))

		_, err := repo.FindActiveByID(ctx, snippet.ID, snippet.Expires.Add(-time.Second))
		assert.NoError(t, err)

		_, err = repo.FindActiveByID(ctx, snippet.ID, snippet.Expires)
		assert.ErrorIs(t, err, apperror.ErrNotFound)

		_, err = repo.FindActiveByID(ctx, snippet.ID, snippet.Expires.Add(time.Second))
		assert.ErrorIs(t, err, apperror.ErrNotFound)
	})

	t.Run("unknown ids are not found", func(t *testing.T) {
		repo := repository.NewSnippetRepository(newDB(t))

		_, err := repo.FindActiveByID(ctx, 1, testutil.Epoch)

		assert.ErrorIs(t, err, apperror.ErrNotFound)
	})

	t.Run("reads times back in UTC even when they were stored in another zone", func(t *testing.T) {
		repo := repository.NewSnippetRepository(newDB(t))
		zone := time.FixedZone("WIB", 7*3600)
		snippet := newSnippet("zoned", 7)
		snippet.Created = snippet.Created.In(zone)
		snippet.Expires = snippet.Expires.In(zone)
		require.NoError(t, repo.Create(ctx, snippet))

		found, err := repo.FindActiveByID(ctx, snippet.ID, testutil.Epoch)
		require.NoError(t, err)
		listed, err := repo.ListLatestActive(ctx, testutil.Epoch, 10)
		require.NoError(t, err)
		require.Len(t, listed, 1)

		for _, got := range []model.Snippet{*found, listed[0]} {
			assert.Equal(t, time.UTC, got.Created.Location())
			assert.Equal(t, time.UTC, got.Expires.Location())
			assert.True(t, snippet.Expires.Equal(got.Expires))
		}
	})

	t.Run("accepts times in any zone", func(t *testing.T) {
		repo := repository.NewSnippetRepository(newDB(t))
		snippet := newSnippet("zoned", 1)
		require.NoError(t, repo.Create(ctx, snippet))

		_, err := repo.FindActiveByID(ctx, snippet.ID, testutil.Epoch.In(time.FixedZone("WIB", 7*3600)))

		assert.NoError(t, err)
	})
}

func stringOfLen(n int) string {
	b := make([]byte, n)
	for i := range b {
		b[i] = 'x'
	}
	return string(b)
}

func TestSnippetRepository_ListLatestActive(t *testing.T) {
	ctx := context.Background()

	t.Run("returns an empty non-nil slice when there is nothing", func(t *testing.T) {
		repo := repository.NewSnippetRepository(newDB(t))

		list, err := repo.ListLatestActive(ctx, testutil.Epoch, 10)

		require.NoError(t, err)
		assert.NotNil(t, list)
		assert.Empty(t, list)
	})

	t.Run("orders newest first, respects the limit and skips expired rows", func(t *testing.T) {
		repo := repository.NewSnippetRepository(newDB(t))
		require.NoError(t, repo.Create(ctx, newSnippet("expired", 1)))
		for _, title := range []string{"a", "b", "c", "d"} {
			require.NoError(t, repo.Create(ctx, newSnippet(title, 7)))
		}
		now := testutil.Epoch.Add(2 * 24 * time.Hour)

		list, err := repo.ListLatestActive(ctx, now, 3)

		require.NoError(t, err)
		require.Len(t, list, 3)
		assert.Equal(t, []string{"d", "c", "b"}, []string{list[0].Title, list[1].Title, list[2].Title})
		assert.Equal(t, time.UTC, list[0].Created.Location())

		all, err := repo.ListLatestActive(ctx, now, 10)
		require.NoError(t, err)
		assert.Len(t, all, 4)
	})
}

func TestSnippetRepository_ReportsDatabaseFailures(t *testing.T) {
	ctx := context.Background()
	db := newDB(t)
	repo := repository.NewSnippetRepository(db)
	require.NoError(t, database.Close(db))

	err := repo.Create(ctx, newSnippet("x", 1))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "repository: create snippet")

	_, err = repo.FindActiveByID(ctx, 1, testutil.Epoch)
	require.Error(t, err)
	assert.NotErrorIs(t, err, apperror.ErrNotFound)

	_, err = repo.ListLatestActive(ctx, testutil.Epoch, 10)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "repository: list latest snippets")
}
