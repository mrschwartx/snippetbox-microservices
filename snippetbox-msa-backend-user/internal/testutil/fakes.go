package testutil

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"snippetbox/backend-user/internal/apperror"
	"snippetbox/backend-user/internal/model"
)

type FakeHasher struct{}

func (FakeHasher) Hash(plain string) (string, error) {
	return "hashed:" + plain, nil
}

func (FakeHasher) Verify(plain, hash string) bool {
	return hash == "hashed:"+plain
}

type issuedToken struct {
	userID    uint
	expiresAt time.Time
}

type FakeTokens struct {
	mu     sync.Mutex
	ttl    time.Duration
	issued map[string]issuedToken
}

func NewFakeTokens(ttl time.Duration) *FakeTokens {
	return &FakeTokens{ttl: ttl, issued: make(map[string]issuedToken)}
}

func (f *FakeTokens) Issue(userID uint, now time.Time) (model.Token, error) {
	f.mu.Lock()
	defer f.mu.Unlock()

	value := fmt.Sprintf("token-%d-%d", userID, len(f.issued)+1)
	expiresAt := now.Add(f.ttl)
	f.issued[value] = issuedToken{userID: userID, expiresAt: expiresAt}

	return model.Token{Value: value, ExpiresAt: expiresAt}, nil
}

func (f *FakeTokens) Verify(token string, now time.Time) (uint, error) {
	f.mu.Lock()
	defer f.mu.Unlock()

	issued, ok := f.issued[token]
	if !ok || !now.Before(issued.expiresAt) {
		return 0, apperror.ErrUnauthenticated
	}
	return issued.userID, nil
}

type MemUserRepo struct {
	mu    sync.Mutex
	users map[uint]model.User
	next  uint
}

func NewMemUserRepo() *MemUserRepo {
	return &MemUserRepo{users: make(map[uint]model.User), next: 1}
}

func (r *MemUserRepo) Create(_ context.Context, user *model.User) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	for _, existing := range r.users {
		if strings.EqualFold(existing.Email, user.Email) {
			return apperror.ErrDuplicateEmail
		}
	}

	user.ID = r.next
	r.next++
	r.users[user.ID] = *user
	return nil
}

func (r *MemUserRepo) FindByEmail(_ context.Context, email string) (*model.User, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	for _, user := range r.users {
		if strings.EqualFold(user.Email, email) {
			found := user
			return &found, nil
		}
	}
	return nil, apperror.ErrNotFound
}

func (r *MemUserRepo) FindByID(_ context.Context, id uint) (*model.User, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	user, ok := r.users[id]
	if !ok {
		return nil, apperror.ErrNotFound
	}
	return &user, nil
}

func (r *MemUserRepo) Count() int {
	r.mu.Lock()
	defer r.mu.Unlock()
	return len(r.users)
}
