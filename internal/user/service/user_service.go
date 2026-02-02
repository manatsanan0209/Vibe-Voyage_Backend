package service

import (
	"context"
	"errors"
	"github.com/manatsanan0209/Vibe-Voyage_Backend/internal/domain"
)

type userService struct {
	repo domain.UserRepository
}

func NewUserService(repo domain.UserRepository) domain.UserService {
	return &userService{repo: repo}
}

func (s *userService) Register(ctx context.Context, user *domain.User) error {
	if user.Email == "" {
		return errors.New("email is required")
	}
	// ... hash password, validate logic, etc. ...
	return s.repo.Create(ctx, user)
}

func (s *userService) FindUser(ctx context.Context, id int) (*domain.User, error) {
	return s.repo.GetByID(ctx, id)
}