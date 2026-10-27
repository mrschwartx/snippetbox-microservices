package model

import (
	"strings"
	"time"
)

const (
	PasswordMinLength = 8
	PasswordMaxBytes  = 72
	NameMaxLength     = 255
	EmailMaxLength    = 255
)

type User struct {
	ID             uint      `gorm:"primaryKey"`
	Name           string    `gorm:"size:255;not null"`
	Email          string    `gorm:"size:255;not null;uniqueIndex:users_uc_email"`
	HashedPassword string    `gorm:"size:255;not null"`
	Created        time.Time `gorm:"not null"`
}

func (User) TableName() string {
	return "users"
}

func NormalizeEmail(email string) string {
	return strings.ToLower(strings.TrimSpace(email))
}
