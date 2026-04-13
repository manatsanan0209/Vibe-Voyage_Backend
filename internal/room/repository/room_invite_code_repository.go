package repository

import (
	"context"
	"errors"

	"github.com/manatsanan0209/Vibe-Voyage_Backend/internal/domain"
	"gorm.io/gorm"
)

func (r *roomRepository) Create(ctx context.Context, invite *domain.RoomInviteCode) error {
	return r.db.WithContext(ctx).Create(invite).Error
}

func (r *roomRepository) GetByCode(ctx context.Context, code string) (*domain.RoomInviteCode, error) {
	var invite domain.RoomInviteCode
	if err := r.db.WithContext(ctx).Where("invite_code = ?", code).First(&invite).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &invite, nil
}

func (r *roomRepository) ListByRoomID(ctx context.Context, roomID uint) ([]domain.RoomInviteCode, error) {
	var invites []domain.RoomInviteCode
	if err := r.db.WithContext(ctx).
		Where("room_id = ?", roomID).
		Order("created_at DESC").
		Find(&invites).Error; err != nil {
		return nil, err
	}
	return invites, nil
}

func (r *roomRepository) ExistsCode(ctx context.Context, code string) (bool, error) {
	var count int64
	if err := r.db.WithContext(ctx).Model(&domain.RoomInviteCode{}).
		Where("invite_code = ?", code).
		Count(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
}
