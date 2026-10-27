package security_test

import (
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"

	"snippetbox/backend-user/internal/apperror"
	"snippetbox/backend-user/internal/security"
)

const secret = "0123456789abcdef0123456789abcdef"

var now = time.Date(2026, 10, 15, 8, 0, 0, 0, time.UTC)

func TestBcryptHasher(t *testing.T) {
	t.Parallel()

	hasher := security.NewBcryptHasher(bcrypt.MinCost)

	hash, err := hasher.Hash("pa55word")
	require.NoError(t, err)
	assert.NotEqual(t, "pa55word", hash)

	assert.True(t, hasher.Verify("pa55word", hash))
	assert.False(t, hasher.Verify("wrong", hash))
	assert.False(t, hasher.Verify("pa55word", "not-a-hash"))

	other, err := hasher.Hash("pa55word")
	require.NoError(t, err)
	assert.NotEqual(t, hash, other)
}

func TestBcryptHasher_RejectsPasswordsOver72Bytes(t *testing.T) {
	t.Parallel()

	_, err := security.NewBcryptHasher(bcrypt.MinCost).Hash(string(make([]byte, 73)))
	assert.Error(t, err)
}

func TestNewBcryptHasher_CostSelection(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		cost int
		want int
	}{
		{"valid cost is kept", bcrypt.MinCost, bcrypt.MinCost},
		{"below minimum falls back to default", bcrypt.MinCost - 1, security.DefaultBcryptCost},
		{"above maximum falls back to default", bcrypt.MaxCost + 1, security.DefaultBcryptCost},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.want, costOf(t, security.NewBcryptHasher(tt.cost)))
		})
	}
}

func costOf(t *testing.T, hasher *security.BcryptHasher) int {
	t.Helper()
	hash, err := hasher.Hash("pa55word")
	require.NoError(t, err)
	cost, err := bcrypt.Cost([]byte(hash))
	require.NoError(t, err)
	return cost
}

func TestJWTService_RoundTrip(t *testing.T) {
	t.Parallel()

	svc := security.NewJWTService(secret, time.Hour)

	token, err := svc.Issue(42, now)
	require.NoError(t, err)
	assert.Equal(t, now.Add(time.Hour), token.ExpiresAt)

	id, err := svc.Verify(token.Value, now.Add(30*time.Minute))
	require.NoError(t, err)
	assert.Equal(t, uint(42), id)
}

func signed(t *testing.T, method jwt.SigningMethod, key any, claims jwt.RegisteredClaims) string {
	t.Helper()
	raw, err := jwt.NewWithClaims(method, claims).SignedString(key)
	require.NoError(t, err)
	return raw
}

func TestJWTService_Verify_Rejects(t *testing.T) {
	t.Parallel()

	svc := security.NewJWTService(secret, time.Hour)
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
			id, err := svc.Verify(tt.token, tt.at)
			assert.ErrorIs(t, err, apperror.ErrUnauthenticated)
			assert.Zero(t, id)
		})
	}
}
