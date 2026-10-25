package handler

import (
	"context"

	"snippetbox/backend-user/internal/dto"
	"snippetbox/backend-user/internal/model"
)

type AuthService interface {
	Signup(ctx context.Context, req dto.SignupRequest) (*model.User, error)
	Login(ctx context.Context, req dto.LoginRequest) (model.Token, *model.User, error)
}
