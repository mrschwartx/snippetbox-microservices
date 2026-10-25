package security

import (
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v5"

	"snippetbox/backend-snippet/internal/apperror"
)

const jwtIssuer = "snippetbox"

type JWTVerifier struct {
	secret []byte
}

func NewJWTVerifier(secret string) *JWTVerifier {
	return &JWTVerifier{secret: []byte(secret)}
}

func (v *JWTVerifier) Verify(raw string, now time.Time) (uint, error) {
	parser := jwt.NewParser(
		jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}),
		jwt.WithIssuer(jwtIssuer),
		jwt.WithExpirationRequired(),
		jwt.WithTimeFunc(func() time.Time { return now }),
	)

	claims := &jwt.RegisteredClaims{}
	token, err := parser.ParseWithClaims(raw, claims, func(*jwt.Token) (any, error) {
		return v.secret, nil
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
