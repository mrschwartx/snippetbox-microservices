//go:build integration

package integration_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"snippetbox/backend-snippet/internal/apperror"
	"snippetbox/backend-snippet/internal/dto"
	"snippetbox/backend-snippet/internal/repository"
	"snippetbox/backend-snippet/internal/security"
	"snippetbox/backend-snippet/internal/service"
	"snippetbox/backend-snippet/internal/testutil"
)

const (
	jwtSecret = "0123456789abcdef0123456789abcdef"
	tokenTTL  = 12 * time.Hour
)

type stack struct {
	clock         *testutil.FixedClock
	authenticator *service.TokenAuthenticator
	snippets      *service.SnippetService
}

func newStack(t *testing.T) *stack {
	t.Helper()

	db := newDB(t)
	clock := testutil.NewFixedClock(testutil.Epoch)

	return &stack{
		clock:         clock,
		authenticator: service.NewTokenAuthenticator(security.NewJWTVerifier(jwtSecret), clock),
		snippets:      service.NewSnippetService(repository.NewSnippetRepository(db), clock),
	}
}

func TestServices_SnippetLifecycle(t *testing.T) {
	ctx := context.Background()
	s := newStack(t)
	token := testutil.IssueToken(t, jwtSecret, 7, s.clock.Now(), tokenTTL)

	actor, err := s.authenticator.Authenticate(ctx, token)
	require.NoError(t, err)
	assert.Equal(t, uint(7), actor.ID)

	short, err := s.snippets.Create(ctx, actor, dto.CreateSnippetRequest{Title: "short", Content: "body", Expires: 1})
	require.NoError(t, err)
	long, err := s.snippets.Create(ctx, actor, dto.CreateSnippetRequest{Title: "long", Content: "body", Expires: 365})
	require.NoError(t, err)

	got, err := s.snippets.Get(ctx, short.ID)
	require.NoError(t, err)
	assert.Equal(t, "short", got.Title)

	list, err := s.snippets.ListLatest(ctx)
	require.NoError(t, err)
	require.Len(t, list, 2)
	assert.Equal(t, long.ID, list[0].ID)

	s.clock.Advance(23 * time.Hour)
	_, err = s.snippets.Get(ctx, short.ID)
	assert.NoError(t, err)

	s.clock.Advance(time.Hour)
	_, err = s.snippets.Get(ctx, short.ID)
	assert.ErrorIs(t, err, apperror.ErrNotFound)
	list, err = s.snippets.ListLatest(ctx)
	require.NoError(t, err)
	require.Len(t, list, 1)
	assert.Equal(t, "long", list[0].Title)
}

func TestServices_TokensExpireOnTheServiceClock(t *testing.T) {
	ctx := context.Background()
	s := newStack(t)
	token := testutil.IssueToken(t, jwtSecret, 7, s.clock.Now(), tokenTTL)

	s.clock.Advance(tokenTTL - time.Second)
	_, err := s.authenticator.Authenticate(ctx, token)
	assert.NoError(t, err)

	s.clock.Advance(time.Second)
	_, err = s.authenticator.Authenticate(ctx, token)
	assert.ErrorIs(t, err, apperror.ErrUnauthenticated)
}

func TestServices_ForeignTokensAreRejected(t *testing.T) {
	ctx := context.Background()
	s := newStack(t)
	foreign := testutil.IssueToken(t, "another-secret-another-secret-0000", 7, s.clock.Now(), tokenTTL)

	_, err := s.authenticator.Authenticate(ctx, foreign)

	assert.ErrorIs(t, err, apperror.ErrUnauthenticated)
}
