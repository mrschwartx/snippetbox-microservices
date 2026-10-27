//go:build integration

package integration_test

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"path/filepath"
	"testing"
	"time"

	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"snippetbox/backend-snippet/internal/database"
	"snippetbox/backend-snippet/internal/model"
)

var (
	quiet       = slog.New(slog.NewTextHandler(io.Discard, nil))
	errNotYet   = errors.New("database not ready")
	immediately = database.Retry{Attempts: 1}
)

type flakyDialector struct {
	gorm.Dialector
	failures int
	calls    int
}

func (d *flakyDialector) Initialize(db *gorm.DB) error {
	d.calls++
	if d.calls <= d.failures {
		return errNotYet
	}
	return d.Dialector.Initialize(db)
}

func fileDialector(t *testing.T) gorm.Dialector {
	t.Helper()
	return sqlite.Open(filepath.Join(t.TempDir(), "snippetbox.db"))
}

func TestConnect_SucceedsAndConfiguresPool(t *testing.T) {
	db, err := database.Connect(context.Background(), fileDialector(t), immediately, quiet)
	require.NoError(t, err)
	t.Cleanup(func() { _ = database.Close(db) })

	sqlDB, err := db.DB()
	require.NoError(t, err)
	assert.Equal(t, 25, sqlDB.Stats().MaxOpenConnections)
	assert.NoError(t, sqlDB.Ping())
}

func TestConnect_RetriesUntilTheDatabaseIsReady(t *testing.T) {
	dialector := &flakyDialector{Dialector: fileDialector(t), failures: 2}

	db, err := database.Connect(context.Background(), dialector, database.Retry{Attempts: 5, Delay: time.Millisecond}, quiet)

	require.NoError(t, err)
	t.Cleanup(func() { _ = database.Close(db) })
	assert.Equal(t, 3, dialector.calls)
}

func TestConnect_GivesUpAfterTheConfiguredAttempts(t *testing.T) {
	dialector := &flakyDialector{Dialector: fileDialector(t), failures: 100}

	db, err := database.Connect(context.Background(), dialector, database.Retry{Attempts: 3, Delay: time.Millisecond}, quiet)

	assert.Nil(t, db)
	require.ErrorIs(t, err, errNotYet)
	assert.Contains(t, err.Error(), "after 3 attempts")
	assert.Equal(t, 3, dialector.calls)
}

func TestConnect_TreatsNonPositiveAttemptsAsOne(t *testing.T) {
	dialector := &flakyDialector{Dialector: fileDialector(t), failures: 100}

	_, err := database.Connect(context.Background(), dialector, database.Retry{Attempts: 0}, quiet)

	require.ErrorIs(t, err, errNotYet)
	assert.Equal(t, 1, dialector.calls)
	assert.Contains(t, err.Error(), "after 1 attempts")
}

func TestConnect_DoesNotWaitAfterTheLastAttempt(t *testing.T) {
	dialector := &flakyDialector{Dialector: fileDialector(t), failures: 100}
	start := time.Now()

	_, err := database.Connect(context.Background(), dialector, database.Retry{Attempts: 1, Delay: time.Hour}, quiet)

	require.Error(t, err)
	assert.Less(t, time.Since(start), 5*time.Second)
}

func TestConnect_StopsWaitingWhenTheContextIsCancelled(t *testing.T) {
	dialector := &flakyDialector{Dialector: fileDialector(t), failures: 100}
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()
	start := time.Now()

	_, err := database.Connect(ctx, dialector, database.Retry{Attempts: 10, Delay: time.Hour}, quiet)

	assert.ErrorIs(t, err, context.DeadlineExceeded)
	assert.Less(t, time.Since(start), 5*time.Second)
	assert.Equal(t, 1, dialector.calls)
}

func TestMigrate_CreatesTablesAndIndexesIdempotently(t *testing.T) {
	db := newDB(t)
	migrator := db.Migrator()

	assert.True(t, migrator.HasTable(&model.Snippet{}))
	assert.True(t, migrator.HasIndex(&model.Snippet{}, "idx_snippets_created"))
	assert.True(t, migrator.HasIndex(&model.Snippet{}, "idx_snippets_expires"))

	assert.NoError(t, database.Migrate(db))
}

func TestMigrate_ReportsFailures(t *testing.T) {
	db := newDB(t)
	require.NoError(t, database.Close(db))

	err := database.Migrate(db)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "database: migrate")
}

func TestClose_ReleasesTheConnectionPool(t *testing.T) {
	db := newDB(t)

	require.NoError(t, database.Close(db))

	sqlDB, err := db.DB()
	require.NoError(t, err)
	assert.Error(t, sqlDB.Ping())
}
