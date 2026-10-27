//go:build integration

package integration_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"snippetbox/backend-user/internal/apperror"
	"snippetbox/backend-user/internal/database"
	"snippetbox/backend-user/internal/model"
	"snippetbox/backend-user/internal/repository"
	"snippetbox/backend-user/internal/testutil"
)

func newDB(t *testing.T) *gorm.DB {
	t.Helper()
	return testutil.NewSQLiteDB(t)
}

func newUser(email string) *model.User {
	return &model.User{Name: "Ada", Email: email, HashedPassword: "hash", Created: testutil.Epoch}
}

func TestUserRepository_Create(t *testing.T) {
	ctx := context.Background()

	t.Run("assigns an id and persists every column", func(t *testing.T) {
		repo := repository.NewUserRepository(newDB(t))
		user := newUser("ada@example.com")

		require.NoError(t, repo.Create(ctx, user))

		assert.NotZero(t, user.ID)
		found, err := repo.FindByID(ctx, user.ID)
		require.NoError(t, err)
		assert.Equal(t, "Ada", found.Name)
		assert.Equal(t, "ada@example.com", found.Email)
		assert.Equal(t, "hash", found.HashedPassword)
		assert.True(t, testutil.Epoch.Equal(found.Created))
		assert.Equal(t, time.UTC, found.Created.Location())
	})

	t.Run("reads times back in UTC even when they were stored in another zone", func(t *testing.T) {
		repo := repository.NewUserRepository(newDB(t))
		user := newUser("zoned@example.com")
		user.Created = testutil.Epoch.In(time.FixedZone("WIB", 7*3600))
		require.NoError(t, repo.Create(ctx, user))

		byID, err := repo.FindByID(ctx, user.ID)
		require.NoError(t, err)
		byEmail, err := repo.FindByEmail(ctx, "zoned@example.com")
		require.NoError(t, err)

		for _, found := range []*model.User{byID, byEmail} {
			assert.True(t, testutil.Epoch.Equal(found.Created))
			assert.Equal(t, time.UTC, found.Created.Location())
		}
	})

	t.Run("enforces unique emails through the database index", func(t *testing.T) {
		repo := repository.NewUserRepository(newDB(t))
		require.NoError(t, repo.Create(ctx, newUser("ada@example.com")))

		err := repo.Create(ctx, newUser("ada@example.com"))

		assert.ErrorIs(t, err, apperror.ErrDuplicateEmail)
		assert.NoError(t, repo.Create(ctx, newUser("grace@example.com")))
	})

	t.Run("reports unexpected database failures", func(t *testing.T) {
		db := newDB(t)
		repo := repository.NewUserRepository(db)
		require.NoError(t, database.Close(db))

		err := repo.Create(ctx, newUser("ada@example.com"))

		require.Error(t, err)
		assert.NotErrorIs(t, err, apperror.ErrDuplicateEmail)
		assert.Contains(t, err.Error(), "repository: create user")
	})
}

func TestUserRepository_Find(t *testing.T) {
	ctx := context.Background()
	db := newDB(t)
	repo := repository.NewUserRepository(db)
	user := newUser("ada@example.com")
	require.NoError(t, repo.Create(ctx, user))

	t.Run("by email", func(t *testing.T) {
		found, err := repo.FindByEmail(ctx, "ada@example.com")

		require.NoError(t, err)
		assert.Equal(t, user.ID, found.ID)
	})

	t.Run("missing records are reported as not found", func(t *testing.T) {
		_, err := repo.FindByEmail(ctx, "nobody@example.com")
		assert.ErrorIs(t, err, apperror.ErrNotFound)

		_, err = repo.FindByID(ctx, 9999)
		assert.ErrorIs(t, err, apperror.ErrNotFound)
	})

	t.Run("database failures are not disguised as not found", func(t *testing.T) {
		broken := newDB(t)
		brokenRepo := repository.NewUserRepository(broken)
		require.NoError(t, database.Close(broken))

		_, err := brokenRepo.FindByEmail(ctx, "ada@example.com")
		require.Error(t, err)
		assert.NotErrorIs(t, err, apperror.ErrNotFound)

		_, err = brokenRepo.FindByID(ctx, 1)
		require.Error(t, err)
		assert.NotErrorIs(t, err, apperror.ErrNotFound)
	})

	t.Run("a cancelled context aborts the query", func(t *testing.T) {
		cancelled, cancel := context.WithCancel(ctx)
		cancel()

		_, err := repo.FindByEmail(cancelled, "ada@example.com")

		require.Error(t, err)
		assert.NotErrorIs(t, err, apperror.ErrNotFound)
	})
}
