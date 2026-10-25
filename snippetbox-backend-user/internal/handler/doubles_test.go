package handler_test

import (
	"bytes"
	"context"
	"log/slog"

	"snippetbox/backend-user/internal/dto"
	"snippetbox/backend-user/internal/model"
)

type authServiceFake struct {
	signup func(context.Context, dto.SignupRequest) (*model.User, error)
	login  func(context.Context, dto.LoginRequest) (model.Token, *model.User, error)
}

func (f authServiceFake) Signup(ctx context.Context, req dto.SignupRequest) (*model.User, error) {
	return f.signup(ctx, req)
}

func (f authServiceFake) Login(ctx context.Context, req dto.LoginRequest) (model.Token, *model.User, error) {
	return f.login(ctx, req)
}

func newLogger() (*slog.Logger, *bytes.Buffer) {
	var buf bytes.Buffer
	return slog.New(slog.NewJSONHandler(&buf, nil)), &buf
}
