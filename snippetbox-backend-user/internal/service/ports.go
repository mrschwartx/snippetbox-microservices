package service

import (
	"context"
	"time"

	"snippetbox/backend-user/internal/model"
)

type Clock interface {
	Now() time.Time
}

type UserRepository interface {
	Create(ctx context.Context, user *model.User) error
	FindByEmail(ctx context.Context, email string) (*model.User, error)
	FindByID(ctx context.Context, id uint) (*model.User, error)
}

type PasswordHasher interface {
	Hash(plain string) (string, error)
	Verify(plain, hash string) bool
}

type TokenService interface {
	Issue(userID uint, now time.Time) (model.Token, error)
	Verify(token string, now time.Time) (uint, error)
}
