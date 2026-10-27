package service_test

import (
	"context"
	"errors"
	"time"

	"snippetbox/backend-user/internal/model"
	"snippetbox/backend-user/internal/testutil"
)

var errBoom = errors.New("boom")

type userRepoStub struct {
	*testutil.MemUserRepo
	createErr      error
	findByEmailErr error
	findByIDErr    error
}

func (s *userRepoStub) Create(ctx context.Context, user *model.User) error {
	if s.createErr != nil {
		return s.createErr
	}
	return s.MemUserRepo.Create(ctx, user)
}

func (s *userRepoStub) FindByEmail(ctx context.Context, email string) (*model.User, error) {
	if s.findByEmailErr != nil {
		return nil, s.findByEmailErr
	}
	return s.MemUserRepo.FindByEmail(ctx, email)
}

func (s *userRepoStub) FindByID(ctx context.Context, id uint) (*model.User, error) {
	if s.findByIDErr != nil {
		return nil, s.findByIDErr
	}
	return s.MemUserRepo.FindByID(ctx, id)
}

type failingHasher struct{ testutil.FakeHasher }

func (failingHasher) Hash(string) (string, error) { return "", errBoom }

type failingTokens struct{ *testutil.FakeTokens }

func (failingTokens) Issue(uint, time.Time) (model.Token, error) { return model.Token{}, errBoom }
