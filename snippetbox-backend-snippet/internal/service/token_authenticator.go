package service

import (
	"context"

	"snippetbox/backend-snippet/internal/model"
)

type TokenAuthenticator struct {
	tokens TokenVerifier
	clock  Clock
}

func NewTokenAuthenticator(tokens TokenVerifier, clock Clock) *TokenAuthenticator {
	return &TokenAuthenticator{tokens: tokens, clock: clock}
}

func (a *TokenAuthenticator) Authenticate(_ context.Context, token string) (*model.Principal, error) {
	id, err := a.tokens.Verify(token, a.clock.Now())
	if err != nil {
		return nil, err
	}
	return &model.Principal{ID: id}, nil
}
