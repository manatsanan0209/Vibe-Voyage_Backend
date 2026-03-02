package repository

import (
	"context"
	"errors"

	"github.com/manatsanan0209/Vibe-Voyage_Backend/internal/domain"
	"gorm.io/gorm"
)

type userLifestyleRepository struct {
	db *gorm.DB
}

func NewUserLifestyleRepository(db *gorm.DB) domain.UserLifestyleRepository {
	return &userLifestyleRepository{db: db}
}

func (r *userLifestyleRepository) Create(ctx context.Context, lifestyle *domain.UserLifestyle) error {
	return r.db.WithContext(ctx).Create(lifestyle).Error
}

func (r *userLifestyleRepository) GetByID(ctx context.Context, lifestyleID uint) (*domain.UserLifestyle, error) {
	var lifestyle domain.UserLifestyle
	if err := r.db.WithContext(ctx).First(&lifestyle, lifestyleID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &lifestyle, nil
}

func (r *userLifestyleRepository) GetByUserAndRoom(ctx context.Context, userID, roomID uint) (*domain.UserLifestyle, error) {
	var lifestyle domain.UserLifestyle
	if err := r.db.WithContext(ctx).
		Where("user_id = ? AND room_id = ?", userID, roomID).
		First(&lifestyle).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &lifestyle, nil
}

func (r *userLifestyleRepository) GetByRoomID(ctx context.Context, roomID uint) ([]domain.UserLifestyle, error) {
	var lifestyles []domain.UserLifestyle
	if err := r.db.WithContext(ctx).Where("room_id = ?", roomID).Find(&lifestyles).Error; err != nil {
		return nil, err
	}
	return lifestyles, nil
}

func (r *userLifestyleRepository) Update(ctx context.Context, lifestyle *domain.UserLifestyle) error {
	return r.db.WithContext(ctx).Save(lifestyle).Error
}
