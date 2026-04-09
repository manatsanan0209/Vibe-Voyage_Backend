package domain

import (
	"context"
	"time"

	"gorm.io/gorm"
)

const (
	InviteAccessView = "view"
	InviteAccessEdit = "edit"
)

type RoomInviteCode struct {
	RoomInviteID        uint           `json:"room_invite_id" gorm:"primaryKey;autoIncrement"`
	RoomID              uint           `json:"room_id" gorm:"not null;index"`
	Room                Room           `json:"-" gorm:"foreignKey:RoomID;references:RoomID"`
	InviteCodeCreatorID uint           `json:"invite_code_creator_id" gorm:"not null;index"`
	InviteCodeCreator   User           `json:"-" gorm:"foreignKey:InviteCodeCreatorID;references:UserID"`
	InviteCode          string         `json:"invite_code" gorm:"size:32;not null;uniqueIndex"`
	Access              string         `json:"access" gorm:"type:varchar(16);not null;default:view"`
	ExpireTime          time.Time      `json:"expire_time" gorm:"not null;index"`
	CreatedAt           time.Time      `json:"created_at"`
	UpdatedAt           time.Time      `json:"updated_at"`
	DeletedAt           gorm.DeletedAt `json:"deleted_at" gorm:"index"`
}

type RoomInviteCodeRepository interface {
	Create(ctx context.Context, invite *RoomInviteCode) error
	GetByCode(ctx context.Context, code string) (*RoomInviteCode, error)
	ListByRoomID(ctx context.Context, roomID uint) ([]RoomInviteCode, error)
	ExistsCode(ctx context.Context, code string) (bool, error)
}
