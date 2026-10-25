package security_test

import (
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"snippetbox/backend-snippet/internal/apperror"
	"snippetbox/backend-snippet/internal/security"
	"snippetbox/backend-snippet/internal/testutil"
)

const secret = "0123456789abcdef0123456789abcdef"

var now = time.Date(2026, 10, 15, 8, 0, 0, 0, time.UTC)

func signed(t *testing.T, method jwt.SigningMethod, key any, claims jwt.RegisteredClaims) string {
	t.Helper()
	raw, err := jwt.NewWithClaims(method, claims).SignedString(key)
	require.NoError(t, err)
	return raw
}

func TestJWTVerifier_AcceptsTokensIssuedByTheUserService(t *testing.T) {
	t.Parallel()

	verifier := security.NewJWTVerifier(secret)
	token := testutil.IssueToken(t, secret, 42, now, time.Hour)

	id, err := verifier.Verify(token, now.Add(30*time.Minute))

	require.NoError(t, err)
	assert.Equal(t, uint(42), id)
}

func TestJWTVerifier_Verify_Rejects(t *testing.T) {
	t.Parallel()

	verifier := security.NewJWTVerifier(secret)
	valid := jwt.RegisteredClaims{
		Issuer:    "snippetbox",
		Subject:   "7",
		IssuedAt:  jwt.NewNumericDate(now),
		ExpiresAt: jwt.NewNumericDate(now.Add(time.Hour)),
	}

	with := func(mutate func(*jwt.RegisteredClaims)) jwt.RegisteredClaims {
		claims := valid
		mutate(&claims)
		return claims
	}

	tests := []struct {
		name  string
		token string
		at    time.Time
	}{
		{"expired", signed(t, jwt.SigningMethodHS256, []byte(secret), valid), now.Add(2 * time.Hour)},
		{"wrong secret", signed(t, jwt.SigningMethodHS256, []byte("another-secret-another-secret-0000"), valid), now},
		{"wrong issuer", signed(t, jwt.SigningMethodHS256, []byte(secret), with(func(c *jwt.RegisteredClaims) { c.Issuer = "evil" })), now},
		{"missing expiry", signed(t, jwt.SigningMethodHS256, []byte(secret), with(func(c *jwt.RegisteredClaims) { c.ExpiresAt = nil })), now},
		{"non numeric subject", signed(t, jwt.SigningMethodHS256, []byte(secret), with(func(c *jwt.RegisteredClaims) { c.Subject = "abc" })), now},
		{"zero subject", signed(t, jwt.SigningMethodHS256, []byte(secret), with(func(c *jwt.RegisteredClaims) { c.Subject = "0" })), now},
		{"other algorithm", signed(t, jwt.SigningMethodHS512, []byte(secret), valid), now},
		{"alg none", signed(t, jwt.SigningMethodNone, jwt.UnsafeAllowNoneSignatureType, valid), now},
		{"garbage", "not.a.jwt", now},
		{"empty", "", now},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			id, err := verifier.Verify(tt.token, tt.at)
			assert.ErrorIs(t, err, apperror.ErrUnauthenticated)
			assert.Zero(t, id)
		})
	}
}
