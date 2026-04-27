package domain

import (
	"context"
	"time"

	"gorm.io/gorm"
)

const (
	NotifTypeRoomInvite        = "room_invite"
	NotifTypeMemberJoined      = "member_joined"
	NotifTypeMemberLeft        = "member_left"
	NotifTypeTripCreated       = "trip_created"
	NotifTypeLifestyleAnalyzed = "lifestyle_analyzed"
	NotifTypeScheduleUpdated   = "schedule_updated"
)

type Notification struct {
	NotificationID uint           `json:"notification_id" gorm:"primaryKey;autoIncrement"`
	UserID         uint           `json:"user_id" gorm:"not null;index"`
	User           User           `json:"-" gorm:"foreignKey:UserID;references:UserID"`
	Type           string         `json:"type" gorm:"not null"`
	Title          string         `json:"title" gorm:"not null"`
	Message        string         `json:"message" gorm:"not null"`
	IsRead         bool           `json:"is_read" gorm:"default:false"`
	ReferenceID    *uint          `json:"reference_id,omitempty"`
	ReferenceType  *string        `json:"reference_type,omitempty"`
	CreatedAt      time.Time      `json:"created_at"`
	UpdatedAt      time.Time      `json:"updated_at"`
	DeletedAt      gorm.DeletedAt `json:"-" gorm:"index"`
}

type NotificationRepository interface {
	Create(ctx context.Context, n *Notification) error
	GetByUserID(ctx context.Context, userID uint, unreadOnly bool) ([]Notification, error)
	MarkAsRead(ctx context.Context, notificationID, userID uint) error
	MarkAllAsRead(ctx context.Context, userID uint) error
	Delete(ctx context.Context, notificationID, userID uint) error
}

type NotificationService interface {
	GetNotifications(ctx context.Context, userID uint, unreadOnly bool) ([]Notification, error)
	MarkAsRead(ctx context.Context, notificationID, userID uint) error
	MarkAllAsRead(ctx context.Context, userID uint) error
	Delete(ctx context.Context, notificationID, userID uint) error
	Notify(ctx context.Context, userID uint, notifType, title, message string, refID *uint, refType *string) error
}
