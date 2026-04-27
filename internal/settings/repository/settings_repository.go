package repository

import (
	"context"

	"github.com/manatsanan0209/Vibe-Voyage_Backend/internal/domain"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type pgUserSettingsRepository struct {
	db *gorm.DB
}

func NewUserSettingsRepository(db *gorm.DB) domain.UserSettingsRepository {
	return &pgUserSettingsRepository{db: db}
}

func (r *pgUserSettingsRepository) GetByUserID(ctx context.Context, userID uint) (*domain.UserSettings, error) {
	var settings domain.UserSettings
	err := r.db.WithContext(ctx).Where("user_id = ?", userID).First(&settings).Error
	if err != nil {
		return nil, err
	}
	return &settings, nil
}

func (r *pgUserSettingsRepository) Upsert(ctx context.Context, settings *domain.UserSettings) error {
	return r.db.WithContext(ctx).
		Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "user_id"}},
			DoUpdates: clause.AssignmentColumns([]string{"theme", "language", "date_format", "time_format", "notify_room_invite", "notify_member_joined", "notify_member_left", "notify_trip_created", "notify_lifestyle_analyzed", "notify_schedule_updated", "updated_at"}),
		}).
		Create(settings).Error
}
