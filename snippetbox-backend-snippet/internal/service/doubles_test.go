package service_test

import (
	"context"
	"errors"
	"time"

	"snippetbox/backend-snippet/internal/model"
	"snippetbox/backend-snippet/internal/testutil"
)

var errBoom = errors.New("boom")

type snippetRepoStub struct {
	*testutil.MemSnippetRepo
	err error
}

func (s *snippetRepoStub) Create(ctx context.Context, snippet *model.Snippet) error {
	if s.err != nil {
		return s.err
	}
	return s.MemSnippetRepo.Create(ctx, snippet)
}

func (s *snippetRepoStub) FindActiveByID(ctx context.Context, id uint, now time.Time) (*model.Snippet, error) {
	if s.err != nil {
		return nil, s.err
	}
	return s.MemSnippetRepo.FindActiveByID(ctx, id, now)
}

func (s *snippetRepoStub) ListLatestActive(ctx context.Context, now time.Time, limit int) ([]model.Snippet, error) {
	if s.err != nil {
		return nil, s.err
	}
	return s.MemSnippetRepo.ListLatestActive(ctx, now, limit)
}
