package dto

import (
	"strings"
	"time"

	"snippetbox/backend-user/internal/model"
	"snippetbox/backend-user/internal/validator"
)

type SignupRequest struct {
	Name     string `json:"name"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (r SignupRequest) Validate() error {
	var v validator.Validator

	v.CheckField(validator.NotBlank(r.Name), "name", messageBlank)
	v.CheckField(validator.MaxChars(r.Name, model.NameMaxLength), "name", messageNameTooLong)

	v.CheckField(validator.NotBlank(r.Email), "email", messageBlank)
	v.CheckField(validator.Matches(strings.TrimSpace(r.Email), validator.EmailRX), "email", messageInvalidEmail)
	v.CheckField(validator.MaxChars(r.Email, model.EmailMaxLength), "email", messageEmailTooLong)

	v.CheckField(validator.NotBlank(r.Password), "password", messageBlank)
	v.CheckField(validator.MinChars(r.Password, model.PasswordMinLength), "password", messagePasswordShort)
	v.CheckField(validator.MaxBytes(r.Password, model.PasswordMaxBytes), "password", messagePasswordLong)

	return v.Err()
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (r LoginRequest) Validate() error {
	var v validator.Validator

	v.CheckField(validator.NotBlank(r.Email), "email", messageBlank)
	v.CheckField(validator.Matches(strings.TrimSpace(r.Email), validator.EmailRX), "email", messageInvalidEmail)
	v.CheckField(validator.NotBlank(r.Password), "password", messageBlank)

	return v.Err()
}

type UserResponse struct {
	ID      uint      `json:"id"`
	Name    string    `json:"name"`
	Email   string    `json:"email"`
	Created time.Time `json:"created"`
}

type LoginResponse struct {
	Token     string       `json:"token"`
	TokenType string       `json:"token_type"`
	ExpiresAt time.Time    `json:"expires_at"`
	User      UserResponse `json:"user"`
}

func NewUserResponse(u *model.User) UserResponse {
	return UserResponse{
		ID:      u.ID,
		Name:    u.Name,
		Email:   u.Email,
		Created: u.Created.UTC(),
	}
}

func NewLoginResponse(token model.Token, u *model.User) LoginResponse {
	return LoginResponse{
		Token:     token.Value,
		TokenType: "Bearer",
		ExpiresAt: token.ExpiresAt.UTC(),
		User:      NewUserResponse(u),
	}
}
