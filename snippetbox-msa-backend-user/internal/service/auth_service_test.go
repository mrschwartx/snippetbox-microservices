package service_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"snippetbox/backend-user/internal/apperror"
	"snippetbox/backend-user/internal/dto"
	"snippetbox/backend-user/internal/service"
	"snippetbox/backend-user/internal/testutil"
)

type authFixture struct {
	svc    *service.AuthService
	users  *userRepoStub
	clock  *testutil.FixedClock
	tokens *testutil.FakeTokens
}

func newAuthFixture() *authFixture {
	f := &authFixture{
		users:  &userRepoStub{MemUserRepo: testutil.NewMemUserRepo()},
		clock:  testutil.NewFixedClock(testutil.Epoch),
		tokens: testutil.NewFakeTokens(time.Hour),
	}
	f.svc = service.NewAuthService(f.users, testutil.FakeHasher{}, f.tokens, f.clock)
	return f
}

var signupReq = dto.SignupRequest{Name: "  Ada Lovelace ", Email: " Ada@Example.com ", Password: "pa55word"}

func TestAuthService_Signup(t *testing.T) {
	t.Parallel()

	t.Run("creates a normalized user", func(t *testing.T) {
		t.Parallel()
		f := newAuthFixture()

		user, err := f.svc.Signup(context.Background(), signupReq)

		require.NoError(t, err)
		assert.NotZero(t, user.ID)
		assert.Equal(t, "Ada Lovelace", user.Name)
		assert.Equal(t, "ada@example.com", user.Email)
		assert.Equal(t, "hashed:pa55word", user.HashedPassword)
		assert.Equal(t, testutil.Epoch, user.Created)
	})

	t.Run("rejects invalid input before touching the repository", func(t *testing.T) {
		t.Parallel()
		f := newAuthFixture()
		f.users.findByEmailErr = errBoom

		_, err := f.svc.Signup(context.Background(), dto.SignupRequest{})

		var verr *apperror.ValidationError
		require.ErrorAs(t, err, &verr)
		assert.Contains(t, verr.FieldErrors, "name")
		assert.Zero(t, f.users.Count())
	})

	t.Run("rejects an email that is already registered", func(t *testing.T) {
		t.Parallel()
		f := newAuthFixture()
		_, err := f.svc.Signup(context.Background(), signupReq)
		require.NoError(t, err)

		again := signupReq
		again.Email = "ADA@example.com"
		_, err = f.svc.Signup(context.Background(), again)

		var verr *apperror.ValidationError
		require.ErrorAs(t, err, &verr)
		assert.Equal(t, "Email address is already in use", verr.FieldErrors["email"])
		assert.Equal(t, 1, f.users.Count())
	})

	t.Run("maps a duplicate detected at insert time to a validation error", func(t *testing.T) {
		t.Parallel()
		f := newAuthFixture()
		f.users.createErr = apperror.ErrDuplicateEmail

		_, err := f.svc.Signup(context.Background(), signupReq)

		var verr *apperror.ValidationError
		require.ErrorAs(t, err, &verr)
		assert.Contains(t, verr.FieldErrors, "email")
	})

	t.Run("propagates unexpected failures with context", func(t *testing.T) {
		t.Parallel()

		tests := []struct {
			name   string
			setup  func(*authFixture) *service.AuthService
			phrase string
		}{
			{"lookup", func(f *authFixture) *service.AuthService {
				f.users.findByEmailErr = errBoom
				return f.svc
			}, "find user"},
			{"hashing", func(f *authFixture) *service.AuthService {
				return service.NewAuthService(f.users, failingHasher{}, f.tokens, f.clock)
			}, "hash password"},
			{"insert", func(f *authFixture) *service.AuthService {
				f.users.createErr = errBoom
				return f.svc
			}, "create user"},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				t.Parallel()
				f := newAuthFixture()
				svc := tt.setup(f)

				_, err := svc.Signup(context.Background(), signupReq)

				require.ErrorIs(t, err, errBoom)
				assert.Contains(t, err.Error(), tt.phrase)
			})
		}
	})
}

func TestAuthService_Login(t *testing.T) {
	t.Parallel()

	registered := func(t *testing.T) *authFixture {
		t.Helper()
		f := newAuthFixture()
		_, err := f.svc.Signup(context.Background(), signupReq)
		require.NoError(t, err)
		return f
	}

	t.Run("returns a token for valid credentials", func(t *testing.T) {
		t.Parallel()
		f := registered(t)

		token, user, err := f.svc.Login(context.Background(), dto.LoginRequest{Email: "ADA@example.com", Password: "pa55word"})

		require.NoError(t, err)
		assert.NotEmpty(t, token.Value)
		assert.Equal(t, testutil.Epoch.Add(time.Hour), token.ExpiresAt)
		assert.Equal(t, "ada@example.com", user.Email)
	})

	t.Run("hides whether the account exists", func(t *testing.T) {
		t.Parallel()
		f := registered(t)

		_, _, unknown := f.svc.Login(context.Background(), dto.LoginRequest{Email: "nobody@example.com", Password: "pa55word"})
		_, _, wrong := f.svc.Login(context.Background(), dto.LoginRequest{Email: "ada@example.com", Password: "wrong-pass"})

		assert.ErrorIs(t, unknown, apperror.ErrInvalidCredentials)
		assert.ErrorIs(t, wrong, apperror.ErrInvalidCredentials)
	})

	t.Run("validates input", func(t *testing.T) {
		t.Parallel()
		f := registered(t)

		_, _, err := f.svc.Login(context.Background(), dto.LoginRequest{})

		var verr *apperror.ValidationError
		assert.ErrorAs(t, err, &verr)
	})

	t.Run("propagates unexpected failures", func(t *testing.T) {
		t.Parallel()

		f := registered(t)
		f.users.findByEmailErr = errBoom
		_, _, err := f.svc.Login(context.Background(), dto.LoginRequest{Email: "ada@example.com", Password: "pa55word"})
		assert.ErrorIs(t, err, errBoom)

		g := registered(t)
		svc := service.NewAuthService(g.users, testutil.FakeHasher{}, failingTokens{g.tokens}, g.clock)
		_, _, err = svc.Login(context.Background(), dto.LoginRequest{Email: "ada@example.com", Password: "pa55word"})
		assert.ErrorIs(t, err, errBoom)
		assert.Contains(t, err.Error(), "issue token")
	})
}

func TestAuthService_Authenticate(t *testing.T) {
	t.Parallel()

	loggedIn := func(t *testing.T) (*authFixture, string) {
		t.Helper()
		f := newAuthFixture()
		_, err := f.svc.Signup(context.Background(), signupReq)
		require.NoError(t, err)
		token, _, err := f.svc.Login(context.Background(), dto.LoginRequest{Email: "ada@example.com", Password: "pa55word"})
		require.NoError(t, err)
		return f, token.Value
	}

	t.Run("resolves the token owner", func(t *testing.T) {
		t.Parallel()
		f, token := loggedIn(t)

		user, err := f.svc.Authenticate(context.Background(), token)

		require.NoError(t, err)
		assert.Equal(t, "ada@example.com", user.Email)
	})

	t.Run("rejects unknown and expired tokens", func(t *testing.T) {
		t.Parallel()
		f, token := loggedIn(t)

		_, err := f.svc.Authenticate(context.Background(), "garbage")
		assert.ErrorIs(t, err, apperror.ErrUnauthenticated)

		f.clock.Advance(2 * time.Hour)
		_, err = f.svc.Authenticate(context.Background(), token)
		assert.ErrorIs(t, err, apperror.ErrUnauthenticated)
	})

	t.Run("rejects a token whose user no longer exists", func(t *testing.T) {
		t.Parallel()
		f, token := loggedIn(t)
		f.users.findByIDErr = apperror.ErrNotFound

		_, err := f.svc.Authenticate(context.Background(), token)

		assert.ErrorIs(t, err, apperror.ErrUnauthenticated)
	})

	t.Run("propagates repository failures", func(t *testing.T) {
		t.Parallel()
		f, token := loggedIn(t)
		f.users.findByIDErr = errBoom

		_, err := f.svc.Authenticate(context.Background(), token)

		require.ErrorIs(t, err, errBoom)
		assert.NotErrorIs(t, err, apperror.ErrUnauthenticated)
	})
}
