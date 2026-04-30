package domain

import (
	"context"
	"time"
)

type UserSettings struct {
	SettingsID uint      `json:"settings_id" gorm:"primaryKey;autoIncrement"`
	UserID     uint      `json:"user_id" gorm:"not null;uniqueIndex"`
	User       User      `json:"-" gorm:"foreignKey:UserID;references:UserID"`
	Theme                  string    `json:"theme" gorm:"default:light"`
	Language               string    `json:"language" gorm:"default:en"`
	DateFormat             string    `json:"date_format" gorm:"default:DD/MM/YYYY"`
	TimeFormat             string    `json:"time_format" gorm:"default:24h"`
	NotifyRoomInvite       bool      `json:"notify_room_invite" gorm:"default:true"`
	NotifyMemberJoined     bool      `json:"notify_member_joined" gorm:"default:true"`
	NotifyMemberLeft       bool      `json:"notify_member_left" gorm:"default:true"`
	NotifyTripCreated      bool      `json:"notify_trip_created" gorm:"default:true"`
	NotifyLifestyleAnalyzed bool     `json:"notify_lifestyle_analyzed" gorm:"default:true"`
	NotifyScheduleUpdated  bool      `json:"notify_schedule_updated" gorm:"default:true"`
	CreatedAt              time.Time `json:"created_at"`
	UpdatedAt              time.Time `json:"updated_at"`
}

type UpdateUserSettingsInput struct {
	Theme                  *string
	Language               *string
	DateFormat             *string
	TimeFormat             *string
	NotifyRoomInvite       *bool
	NotifyMemberJoined     *bool
	NotifyMemberLeft       *bool
	NotifyTripCreated      *bool
	NotifyLifestyleAnalyzed *bool
	NotifyScheduleUpdated  *bool
}

type UserSettingsRepository interface {
	GetByUserID(ctx context.Context, userID uint) (*UserSettings, error)
	Upsert(ctx context.Context, settings *UserSettings) error
}

type UserSettingsService interface {
	GetSettings(ctx context.Context, userID uint) (*UserSettings, error)
	UpdateSettings(ctx context.Context, userID uint, input UpdateUserSettingsInput) (*UserSettings, error)
}
