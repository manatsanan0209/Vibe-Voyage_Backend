package service

import (
	"context"

	"github.com/manatsanan0209/Vibe-Voyage_Backend/internal/domain"
)

type userService struct {
	repo domain.UserRepository
}

func NewUserService(repo domain.UserRepository) domain.UserService {
	return &userService{repo: repo}
}

func (s *userService) CreateUser(ctx context.Context, user *domain.User) error {
	return s.repo.Create(ctx, user)
}

func (s *userService) GetUserByID(ctx context.Context, id uint) (*domain.User, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *userService) GetUserByUsername(ctx context.Context, username string) (*domain.User, error) {
	return s.repo.GetByUsername(ctx, username)
}

func (s *userService) GetUserByEmail(ctx context.Context, email string) (*domain.User, error) {
	return s.repo.GetByEmail(ctx, email)
}