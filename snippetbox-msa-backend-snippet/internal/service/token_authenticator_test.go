package service_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"snippetbox/backend-snippet/internal/apperror"
	"snippetbox/backend-snippet/internal/service"
	"snippetbox/backend-snippet/internal/testutil"
)

type verifierStub struct {
	id       uint
	err      error
	gotToken string
	gotNow   time.Time
}

func (v *verifierStub) Verify(token string, now time.Time) (uint, error) {
	v.gotToken = token
	v.gotNow = now
	return v.id, v.err
}

func TestTokenAuthenticator_Authenticate(t *testing.T) {
	t.Parallel()

	t.Run("resolves the token subject without touching storage", func(t *testing.T) {
		t.Parallel()
		verifier := &verifierStub{id: 7}
		authenticator := service.NewTokenAuthenticator(verifier, testutil.NewFixedClock(testutil.Epoch))

		principal, err := authenticator.Authenticate(context.Background(), "raw-token")

		require.NoError(t, err)
		assert.Equal(t, uint(7), principal.ID)
		assert.Equal(t, "raw-token", verifier.gotToken)
		assert.Equal(t, testutil.Epoch, verifier.gotNow)
	})

	t.Run("rejects tokens the verifier rejects", func(t *testing.T) {
		t.Parallel()
		verifier := &verifierStub{err: apperror.ErrUnauthenticated}
		authenticator := service.NewTokenAuthenticator(verifier, testutil.NewFixedClock(testutil.Epoch))

		principal, err := authenticator.Authenticate(context.Background(), "bad")

		assert.ErrorIs(t, err, apperror.ErrUnauthenticated)
		assert.Nil(t, principal)
	})
}
