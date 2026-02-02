package domain

import "context"

type User struct {
	ID    int    `json:"id"`
	Email string `json:"email"`
	Name  string `json:"name"`
}

type UserRepository interface {
	Create(ctx context.Context, user *User) error
	GetByID(ctx context.Context, id int) (*User, error)
}

type UserService interface {
	Register(ctx context.Context, user *User) error
	FindUser(ctx context.Context, id int) (*User, error)
}