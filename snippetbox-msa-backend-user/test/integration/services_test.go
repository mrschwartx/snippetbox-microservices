//go:build integration

package integration_test

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"

	"snippetbox/backend-user/internal/apperror"
	"snippetbox/backend-user/internal/dto"
	"snippetbox/backend-user/internal/model"
	"snippetbox/backend-user/internal/repository"
	"snippetbox/backend-user/internal/security"
	"snippetbox/backend-user/internal/service"
	"snippetbox/backend-user/internal/testutil"
)

const (
	jwtSecret = "0123456789abcdef0123456789abcdef"
	tokenTTL  = 12 * time.Hour
)

type stack struct {
	db    *gorm.DB
	clock *testutil.FixedClock
	auth  *service.AuthService
}

func newStack(t *testing.T) *stack {
	t.Helper()

	db := newDB(t)
	clock := testutil.NewFixedClock(testutil.Epoch)
	tokens := security.NewJWTService(jwtSecret, tokenTTL)
	hasher := security.NewBcryptHasher(bcrypt.MinCost)

	return &stack{
		db:    db,
		clock: clock,
		auth:  service.NewAuthService(repository.NewUserRepository(db), hasher, tokens, clock),
	}
}

func TestServices_AccountLifecycle(t *testing.T) {
	ctx := context.Background()
	s := newStack(t)

	user, err := s.auth.Signup(ctx, dto.SignupRequest{Name: "Ada", Email: "Ada@Example.com", Password: "pa55word"})
	require.NoError(t, err)
	assert.Equal(t, "ada@example.com", user.Email)
	assert.NotContains(t, user.HashedPassword, "pa55word")

	_, err = s.auth.Signup(ctx, dto.SignupRequest{Name: "Other", Email: "ADA@example.com", Password: "pa55word"})
	var verr *apperror.ValidationError
	require.ErrorAs(t, err, &verr)
	assert.Contains(t, verr.FieldErrors, "email")

	_, _, err = s.auth.Login(ctx, dto.LoginRequest{Email: "ada@example.com", Password: "wrong-password"})
	assert.ErrorIs(t, err, apperror.ErrInvalidCredentials)

	token, loggedIn, err := s.auth.Login(ctx, dto.LoginRequest{Email: "ADA@example.com", Password: "pa55word"})
	require.NoError(t, err)
	assert.Equal(t, user.ID, loggedIn.ID)
	assert.Equal(t, testutil.Epoch.Add(tokenTTL), token.ExpiresAt)

	actor, err := s.auth.Authenticate(ctx, token.Value)
	require.NoError(t, err)
	assert.Equal(t, user.ID, actor.ID)

	s.clock.Advance(tokenTTL - time.Second)
	_, err = s.auth.Authenticate(ctx, token.Value)
	assert.NoError(t, err)

	s.clock.Advance(time.Second)
	_, err = s.auth.Authenticate(ctx, token.Value)
	assert.ErrorIs(t, err, apperror.ErrUnauthenticated)
}

func TestServices_AuthenticateRejectsTokensOfDeletedUsers(t *testing.T) {
	ctx := context.Background()
	s := newStack(t)
	user, err := s.auth.Signup(ctx, dto.SignupRequest{Name: "Ada", Email: "ada@example.com", Password: "pa55word"})
	require.NoError(t, err)
	token, _, err := s.auth.Login(ctx, dto.LoginRequest{Email: "ada@example.com", Password: "pa55word"})
	require.NoError(t, err)

	require.NoError(t, s.db.Delete(&model.User{}, user.ID).Error)

	_, err = s.auth.Authenticate(ctx, token.Value)
	assert.ErrorIs(t, err, apperror.ErrUnauthenticated)
}

func TestServices_ConcurrentSignupsForOneEmailCreateExactlyOneAccount(t *testing.T) {
	ctx := context.Background()
	s := newStack(t)

	const workers = 8
	results := make([]error, workers)
	var wg sync.WaitGroup
	for i := range workers {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, results[i] = s.auth.Signup(ctx, dto.SignupRequest{Name: "Ada", Email: "ada@example.com", Password: "pa55word"})
		}()
	}
	wg.Wait()

	succeeded := 0
	for _, err := range results {
		if err == nil {
			succeeded++
			continue
		}
		var verr *apperror.ValidationError
		require.ErrorAs(t, err, &verr, "losers must see a validation error, never a server error")
		assert.Contains(t, verr.FieldErrors, "email")
	}
	assert.Equal(t, 1, succeeded)

	var count int64
	require.NoError(t, s.db.Model(&model.User{}).Count(&count).Error)
	assert.EqualValues(t, 1, count)
}
