package security

import (
	"fmt"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v5"

	"snippetbox/backend-user/internal/apperror"
	"snippetbox/backend-user/internal/model"
)

const jwtIssuer = "snippetbox"

type JWTService struct {
	secret []byte
	ttl    time.Duration
}

func NewJWTService(secret string, ttl time.Duration) *JWTService {
	return &JWTService{secret: []byte(secret), ttl: ttl}
}

func (s *JWTService) Issue(userID uint, now time.Time) (model.Token, error) {
	expiresAt := now.Add(s.ttl)
	claims := jwt.RegisteredClaims{
		Issuer:    jwtIssuer,
		Subject:   strconv.FormatUint(uint64(userID), 10),
		IssuedAt:  jwt.NewNumericDate(now),
		ExpiresAt: jwt.NewNumericDate(expiresAt),
	}

	signed, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(s.secret)
	if err != nil {
		return model.Token{}, fmt.Errorf("security: sign token: %w", err)
	}

	return model.Token{Value: signed, ExpiresAt: expiresAt.UTC()}, nil
}

func (s *JWTService) Verify(raw string, now time.Time) (uint, error) {
	parser := jwt.NewParser(
		jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}),
		jwt.WithIssuer(jwtIssuer),
		jwt.WithExpirationRequired(),
		jwt.WithTimeFunc(func() time.Time { return now }),
	)

	claims := &jwt.RegisteredClaims{}
	token, err := parser.ParseWithClaims(raw, claims, func(*jwt.Token) (any, error) {
		return s.secret, nil
	})
	if err != nil || !token.Valid {
		return 0, apperror.ErrUnauthenticated
	}

	id, err := strconv.ParseUint(claims.Subject, 10, 64)
	if err != nil || id == 0 {
		return 0, apperror.ErrUnauthenticated
	}

	return uint(id), nil
}
