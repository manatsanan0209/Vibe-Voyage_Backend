package domain

import (
	"context"
	"time"

	"gorm.io/gorm"
)

const (
	RoleOwner  = 1
	RoleMember = 2
)

type RoomMember struct {
	RoomMemberID uint           `json:"room_member_id" gorm:"primaryKey;autoIncrement"`
	RoomID       uint           `json:"room_id" gorm:"not null;index"`
	Room         Room           `json:"-" gorm:"foreignKey:RoomID;references:RoomID"`
	UserID       uint           `json:"user_id" gorm:"not null;index"`
	User         User           `json:"-" gorm:"foreignKey:UserID;references:UserID"`
	Role         int            `json:"role" gorm:"not null"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
	DeletedAt    gorm.DeletedAt `json:"deleted_at" gorm:"index"`
}

type RoomMemberRepository interface {
	GetByRoomID(ctx context.Context, roomID uint) ([]RoomMember, error)
	GetByID(ctx context.Context, roomMemberID uint) (*RoomMember, error)
	AddMember(ctx context.Context, member *RoomMember) (*RoomMember, error)
	DeleteMember(ctx context.Context, roomMemberID uint) error
	ExistsByRoomAndUser(ctx context.Context, roomID, userID uint) (bool, error)
}

type RoomMemberService interface {
	GetMembersByRoomID(ctx context.Context, roomID uint) ([]RoomMember, error)
	AddMember(ctx context.Context, roomID, userID uint) (*RoomMember, error)
	DeleteMember(ctx context.Context, roomID, requesterUserID, roomMemberID uint) error
}
