package service

import (
	"context"
	"time"

	"snippetbox/backend-snippet/internal/model"
)

type Clock interface {
	Now() time.Time
}

type SnippetRepository interface {
	Create(ctx context.Context, snippet *model.Snippet) error
	FindActiveByID(ctx context.Context, id uint, now time.Time) (*model.Snippet, error)
	ListLatestActive(ctx context.Context, now time.Time, limit int) ([]model.Snippet, error)
}

type TokenVerifier interface {
	Verify(token string, now time.Time) (uint, error)
}
