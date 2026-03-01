package domain

import (
	"context"
	"time"

	"gorm.io/gorm"
)

type UserLifestyle struct {
	LifestyleID           uint           `json:"lifestyle_id" gorm:"primaryKey;autoIncrement"`
	UserID                uint           `json:"user_id" gorm:"not null;index"`
	User                  User           `json:"-" gorm:"foreignKey:UserID;references:UserID"`
	RoomID                uint           `json:"room_id" gorm:"not null;index"`
	Room                  Room           `json:"-" gorm:"foreignKey:RoomID;references:RoomID"`
	PreferredDestinations string         `json:"preferred_destinations" gorm:"type:json"`
	VoyagePriorities      string         `json:"voyage_priorities" gorm:"type:json;not null"`
	FoodVibes             string         `json:"food_vibes" gorm:"type:json;not null"`
	AdditionalNotes       string         `json:"additional_notes"`
	StructuredLifestyle   *string        `json:"structured_lifestyle" gorm:"type:json"`
	CreatedAt             time.Time      `json:"created_at"`
	UpdatedAt             time.Time      `json:"updated_at"`
	DeletedAt             gorm.DeletedAt `json:"deleted_at" gorm:"index"`
}

type UserLifestyleRepository interface {
	Create(ctx context.Context, lifestyle *UserLifestyle) error
	GetByUserAndRoom(ctx context.Context, userID, roomID uint) (*UserLifestyle, error)
	GetByRoomID(ctx context.Context, roomID uint) ([]UserLifestyle, error)
	Update(ctx context.Context, lifestyle *UserLifestyle) error
}

type UserLifestyleService interface {
	AnalyzeLifestyle(ctx context.Context, lifestyle *UserLifestyle) error
	GetLifestyle(ctx context.Context, userID, roomID uint) (*UserLifestyle, error)
}
