package repository

import (
	"context"
	"fmt"
	"time"

	"gorm.io/gorm"

	"snippetbox/backend-snippet/internal/model"
)

type SnippetRepository struct {
	db *gorm.DB
}

func NewSnippetRepository(db *gorm.DB) *SnippetRepository {
	return &SnippetRepository{db: db}
}

func (r *SnippetRepository) Create(ctx context.Context, snippet *model.Snippet) error {
	if err := r.db.WithContext(ctx).Create(snippet).Error; err != nil {
		return fmt.Errorf("repository: create snippet: %w", err)
	}
	return nil
}

func (r *SnippetRepository) FindActiveByID(ctx context.Context, id uint, now time.Time) (*model.Snippet, error) {
	var snippet model.Snippet
	err := r.db.WithContext(ctx).
		Where("id = ? AND expires > ?", id, now.UTC()).
		First(&snippet).Error
	if err != nil {
		return nil, findError("find active snippet", err)
	}

	normalizeTimes(&snippet)
	return &snippet, nil
}

func (r *SnippetRepository) ListLatestActive(ctx context.Context, now time.Time, limit int) ([]model.Snippet, error) {
	snippets := []model.Snippet{}
	err := r.db.WithContext(ctx).
		Where("expires > ?", now.UTC()).
		Order("id DESC").
		Limit(limit).
		Find(&snippets).Error
	if err != nil {
		return nil, fmt.Errorf("repository: list latest snippets: %w", err)
	}

	for i := range snippets {
		normalizeTimes(&snippets[i])
	}
	return snippets, nil
}

func normalizeTimes(snippet *model.Snippet) {
	snippet.Created = snippet.Created.UTC()
	snippet.Expires = snippet.Expires.UTC()
}
