package repository

import (
	"context"
	"gorm.io/gorm"
	"github.com/manatsanan0209/Vibe-Voyage_Backend/internal/domain"
)

type pgUserRepository struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) domain.UserRepository {
	return &pgUserRepository{db: db}
}

func (r *pgUserRepository) Create(ctx context.Context, user *domain.User) error {
	err := r.db.WithContext(ctx).Create(user).Error
	if err != nil {
		return err
	}
	return nil
}

func (r *pgUserRepository) GetByID(ctx context.Context, id uint) (*domain.User, error) {
	var user domain.User
	err := r.db.WithContext(ctx).First(&user, id).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *pgUserRepository) GetByUsername(ctx context.Context, username string) (*domain.User, error) {
	var user domain.User
	err := r.db.WithContext(ctx).Where("username = ?", username).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *pgUserRepository) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	var user domain.User
	err := r.db.WithContext(ctx).Where("email = ?", email).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}