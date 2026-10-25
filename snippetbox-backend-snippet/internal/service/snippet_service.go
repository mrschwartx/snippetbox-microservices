package service

import (
	"context"
	"fmt"

	"snippetbox/backend-snippet/internal/apperror"
	"snippetbox/backend-snippet/internal/dto"
	"snippetbox/backend-snippet/internal/model"
)

type SnippetService struct {
	snippets SnippetRepository
	clock    Clock
}

func NewSnippetService(snippets SnippetRepository, clock Clock) *SnippetService {
	return &SnippetService{snippets: snippets, clock: clock}
}

func (s *SnippetService) Create(ctx context.Context, actor *model.Principal, req dto.CreateSnippetRequest) (*model.Snippet, error) {
	if actor == nil {
		return nil, apperror.ErrUnauthenticated
	}
	if err := req.Validate(); err != nil {
		return nil, err
	}

	snippet := model.NewSnippet(req.Title, req.Content, req.Expires, s.clock.Now())
	if err := s.snippets.Create(ctx, &snippet); err != nil {
		return nil, fmt.Errorf("create snippet: %w", err)
	}

	return &snippet, nil
}

func (s *SnippetService) Get(ctx context.Context, id uint) (*model.Snippet, error) {
	if id < 1 {
		return nil, apperror.ErrNotFound
	}
	return s.snippets.FindActiveByID(ctx, id, s.clock.Now())
}

func (s *SnippetService) ListLatest(ctx context.Context) ([]model.Snippet, error) {
	return s.snippets.ListLatestActive(ctx, s.clock.Now(), model.LatestSnippetsLimit)
}
