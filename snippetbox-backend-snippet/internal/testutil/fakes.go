package testutil

import (
	"context"
	"sort"
	"sync"
	"time"

	"snippetbox/backend-snippet/internal/apperror"
	"snippetbox/backend-snippet/internal/model"
)

type MemSnippetRepo struct {
	mu       sync.Mutex
	snippets map[uint]model.Snippet
	next     uint
}

func NewMemSnippetRepo() *MemSnippetRepo {
	return &MemSnippetRepo{snippets: make(map[uint]model.Snippet), next: 1}
}

func (r *MemSnippetRepo) Create(_ context.Context, snippet *model.Snippet) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	snippet.ID = r.next
	r.next++
	r.snippets[snippet.ID] = *snippet
	return nil
}

func (r *MemSnippetRepo) FindActiveByID(_ context.Context, id uint, now time.Time) (*model.Snippet, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	snippet, ok := r.snippets[id]
	if !ok || snippet.IsExpired(now) {
		return nil, apperror.ErrNotFound
	}
	return &snippet, nil
}

func (r *MemSnippetRepo) ListLatestActive(_ context.Context, now time.Time, limit int) ([]model.Snippet, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	active := make([]model.Snippet, 0, len(r.snippets))
	for _, snippet := range r.snippets {
		if !snippet.IsExpired(now) {
			active = append(active, snippet)
		}
	}

	sort.Slice(active, func(i, j int) bool { return active[i].ID > active[j].ID })
	if len(active) > limit {
		active = active[:limit]
	}
	return active, nil
}

func (r *MemSnippetRepo) Count() int {
	r.mu.Lock()
	defer r.mu.Unlock()
	return len(r.snippets)
}
