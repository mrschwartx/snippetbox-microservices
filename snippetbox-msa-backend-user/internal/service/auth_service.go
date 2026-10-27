package service

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"snippetbox/backend-user/internal/apperror"
	"snippetbox/backend-user/internal/dto"
	"snippetbox/backend-user/internal/model"
)

const emailTakenMessage = "Email address is already in use"

type AuthService struct {
	users  UserRepository
	hasher PasswordHasher
	tokens TokenService
	clock  Clock
}

func NewAuthService(users UserRepository, hasher PasswordHasher, tokens TokenService, clock Clock) *AuthService {
	return &AuthService{users: users, hasher: hasher, tokens: tokens, clock: clock}
}

func (s *AuthService) Signup(ctx context.Context, req dto.SignupRequest) (*model.User, error) {
	if err := req.Validate(); err != nil {
		return nil, err
	}

	email := model.NormalizeEmail(req.Email)
	_, err := s.users.FindByEmail(ctx, email)
	switch {
	case err == nil:
		return nil, apperror.NewFieldValidationError("email", emailTakenMessage)
	case !errors.Is(err, apperror.ErrNotFound):
		return nil, fmt.Errorf("signup: find user: %w", err)
	}

	hash, err := s.hasher.Hash(req.Password)
	if err != nil {
		return nil, fmt.Errorf("signup: hash password: %w", err)
	}

	user := &model.User{
		Name:           strings.TrimSpace(req.Name),
		Email:          email,
		HashedPassword: hash,
		Created:        s.clock.Now().UTC(),
	}
	if err := s.users.Create(ctx, user); err != nil {
		if errors.Is(err, apperror.ErrDuplicateEmail) {
			return nil, apperror.NewFieldValidationError("email", emailTakenMessage)
		}
		return nil, fmt.Errorf("signup: create user: %w", err)
	}

	return user, nil
}

func (s *AuthService) Login(ctx context.Context, req dto.LoginRequest) (model.Token, *model.User, error) {
	if err := req.Validate(); err != nil {
		return model.Token{}, nil, err
	}

	user, err := s.users.FindByEmail(ctx, model.NormalizeEmail(req.Email))
	if err != nil {
		if errors.Is(err, apperror.ErrNotFound) {
			return model.Token{}, nil, apperror.ErrInvalidCredentials
		}
		return model.Token{}, nil, fmt.Errorf("login: find user: %w", err)
	}
	if !s.hasher.Verify(req.Password, user.HashedPassword) {
		return model.Token{}, nil, apperror.ErrInvalidCredentials
	}

	token, err := s.tokens.Issue(user.ID, s.clock.Now())
	if err != nil {
		return model.Token{}, nil, fmt.Errorf("login: issue token: %w", err)
	}

	return token, user, nil
}

func (s *AuthService) Authenticate(ctx context.Context, token string) (*model.User, error) {
	id, err := s.tokens.Verify(token, s.clock.Now())
	if err != nil {
		return nil, apperror.ErrUnauthenticated
	}

	user, err := s.users.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, apperror.ErrNotFound) {
			return nil, apperror.ErrUnauthenticated
		}
		return nil, fmt.Errorf("authenticate: find user: %w", err)
	}

	return user, nil
}
