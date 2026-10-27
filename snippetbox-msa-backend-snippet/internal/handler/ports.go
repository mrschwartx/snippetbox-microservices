package handler

import (
	"context"

	"snippetbox/backend-snippet/internal/dto"
	"snippetbox/backend-snippet/internal/model"
)

type SnippetService interface {
	Create(ctx context.Context, actor *model.Principal, req dto.CreateSnippetRequest) (*model.Snippet, error)
	Get(ctx context.Context, id uint) (*model.Snippet, error)
	ListLatest(ctx context.Context) ([]model.Snippet, error)
}
