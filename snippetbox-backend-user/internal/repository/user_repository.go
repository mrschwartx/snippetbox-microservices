package repository

import (
	"context"
	"fmt"

	"gorm.io/gorm"

	"snippetbox/backend-user/internal/apperror"
	"snippetbox/backend-user/internal/model"
)

type UserRepository struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) Create(ctx context.Context, user *model.User) error {
	if err := r.db.WithContext(ctx).Create(user).Error; err != nil {
		if isDuplicateKey(err) {
			return apperror.ErrDuplicateEmail
		}
		return fmt.Errorf("repository: create user: %w", err)
	}
	return nil
}

func (r *UserRepository) FindByEmail(ctx context.Context, email string) (*model.User, error) {
	var user model.User
	if err := r.db.WithContext(ctx).Where("email = ?", email).First(&user).Error; err != nil {
		return nil, findError("find user by email", err)
	}

	user.Created = user.Created.UTC()
	return &user, nil
}

func (r *UserRepository) FindByID(ctx context.Context, id uint) (*model.User, error) {
	var user model.User
	if err := r.db.WithContext(ctx).First(&user, id).Error; err != nil {
		return nil, findError("find user by id", err)
	}

	user.Created = user.Created.UTC()
	return &user, nil
}
