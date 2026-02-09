package domain

import (
	"context"
	"time"

	"gorm.io/gorm"
)

type User struct {
	UserID       uint           `json:"user_id" gorm:"primaryKey;autoIncrement"`
	Username     string         `json:"username" gorm:"unique;not null"`
	Email        string         `json:"email" gorm:"unique;not null"`
	Password     string         `json:"password" gorm:"not null"`
	FullName     string         `json:"full_name" gorm:"not null"`
	ProfileImage string         `json:"profile_image"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
	DeletedAt    gorm.DeletedAt `json:"deleted_at" gorm:"index"`
}

type UserRepository interface {
	Create(ctx context.Context, user *User) error
	GetByID(ctx context.Context, id uint) (*User, error)
	GetByUsername(ctx context.Context, username string) (*User, error)
	GetByEmail(ctx context.Context, email string) (*User, error)
}

type UserService interface {
	CreateUser(ctx context.Context, user *User) error
	GetUserByID(ctx context.Context, id uint) (*User, error)
	GetUserByUsername(ctx context.Context, username string) (*User, error)
	GetUserByEmail(ctx context.Context, email string) (*User, error)
}

type AuthRepository interface {
}

type AuthService interface {
	Register(ctx context.Context, user *User) (*AuthToken, error)
	Login(ctx context.Context, username, password string) (*User, *AuthToken, error)
	ValidateToken(ctx context.Context, token string) (*TokenClaims, error)
}

type AuthToken struct {
	Token     string
	ExpiresAt time.Time
}

type TokenClaims struct {
	UserID    uint
	ExpiresAt time.Time
}
