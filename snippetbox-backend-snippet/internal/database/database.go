package database

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"time"

	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"

	"snippetbox/backend-snippet/internal/model"
)

const (
	maxOpenConns    = 25
	maxIdleConns    = 5
	connMaxLifetime = 30 * time.Minute
)

type Retry struct {
	Attempts int
	Delay    time.Duration
}

func Connect(ctx context.Context, dialector gorm.Dialector, retry Retry, log *slog.Logger) (*gorm.DB, error) {
	attempts := max(retry.Attempts, 1)
	var lastErr error

	for attempt := 1; attempt <= attempts; attempt++ {
		db, err := open(ctx, dialector)
		if err == nil {
			return db, nil
		}

		lastErr = err
		log.Warn("database not ready", "attempt", attempt, "of", attempts, "error", err)

		if attempt == attempts {
			break
		}
		if err := sleep(ctx, retry.Delay); err != nil {
			return nil, err
		}
	}

	return nil, fmt.Errorf("database: connect failed after %d attempts: %w", attempts, lastErr)
}

func Migrate(db *gorm.DB) error {
	if err := db.AutoMigrate(&model.Snippet{}); err != nil {
		return fmt.Errorf("database: migrate: %w", err)
	}
	return nil
}

func Close(db *gorm.DB) error {
	sqlDB, err := db.DB()
	if err != nil {
		return fmt.Errorf("database: resolve connection pool: %w", err)
	}
	return sqlDB.Close()
}

func open(ctx context.Context, dialector gorm.Dialector) (*gorm.DB, error) {
	db, err := gorm.Open(dialector, &gorm.Config{
		TranslateError: true,
		Logger:         gormlogger.Default.LogMode(gormlogger.Warn),
	})
	if err != nil {
		return nil, err
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}
	if err := sqlDB.PingContext(ctx); err != nil {
		_ = sqlDB.Close()
		return nil, err
	}

	configurePool(sqlDB)
	return db, nil
}

func configurePool(sqlDB *sql.DB) {
	sqlDB.SetMaxOpenConns(maxOpenConns)
	sqlDB.SetMaxIdleConns(maxIdleConns)
	sqlDB.SetConnMaxLifetime(connMaxLifetime)
}

func sleep(ctx context.Context, d time.Duration) error {
	timer := time.NewTimer(d)
	defer timer.Stop()

	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return nil
	}
}
