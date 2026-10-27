package security

import (
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

const DefaultBcryptCost = 12

type BcryptHasher struct {
	cost int
}

func NewBcryptHasher(cost int) *BcryptHasher {
	if cost < bcrypt.MinCost || cost > bcrypt.MaxCost {
		cost = DefaultBcryptCost
	}
	return &BcryptHasher{cost: cost}
}

func (h *BcryptHasher) Hash(plain string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(plain), h.cost)
	if err != nil {
		return "", fmt.Errorf("security: hash password: %w", err)
	}
	return string(hash), nil
}

func (h *BcryptHasher) Verify(plain, hash string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(plain)) == nil
}
