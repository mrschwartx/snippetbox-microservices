package testutil

import (
	"strconv"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/require"
)

func IssueToken(t testing.TB, secret string, userID uint, now time.Time, ttl time.Duration) string {
	t.Helper()

	claims := jwt.RegisteredClaims{
		Issuer:    "snippetbox",
		Subject:   strconv.FormatUint(uint64(userID), 10),
		IssuedAt:  jwt.NewNumericDate(now),
		ExpiresAt: jwt.NewNumericDate(now.Add(ttl)),
	}

	raw, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(secret))
	require.NoError(t, err)
	return raw
}
