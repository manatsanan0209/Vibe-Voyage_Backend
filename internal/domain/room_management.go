package domain

import (
	"context"
	"time"

	"gorm.io/gorm"
)

const (
	RoleOwner     = 1
	RoleMember    = 2
	RoleSpectator = 3
)

var RoomRoleMap = map[int]string{
	RoleOwner:     "room_owner",
	RoleMember:    "member",
	RoleSpectator: "spectator",
}

func RoomRoleName(role int) string {
	if name, ok := RoomRoleMap[role]; ok {
		return name
	}
	return "unknown"
}

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

type UserRoomSummary struct {
	RoomID        uint
	TripID        uint
	RoomName      string
	RoomImage     string
	OwnerID       uint
	OwnerUsername string
	Role          int
	JoinedAt      time.Time
	MembersCount  int64
}

type RoomRepository interface {
	GetByRoomID(ctx context.Context, roomID uint) ([]RoomMember, error)
	GetRoomsByUserID(ctx context.Context, userID uint) ([]UserRoomSummary, error)
	GetByID(ctx context.Context, roomMemberID uint) (*RoomMember, error)
	AddMember(ctx context.Context, member *RoomMember) (*RoomMember, error)
	DeleteMember(ctx context.Context, roomMemberID uint) error
	ExistsByRoomAndUser(ctx context.Context, roomID, userID uint) (bool, error)
}

type RoomService interface {
	GetMembersByRoomID(ctx context.Context, roomID uint) ([]RoomMember, error)
	GetRoomsByUserID(ctx context.Context, userID uint) ([]UserRoomSummary, error)
	AddMember(ctx context.Context, roomID, userID uint) (*RoomMember, error)
	DeleteMember(ctx context.Context, roomID, requesterUserID, roomMemberID uint) error
	CreateInviteCode(ctx context.Context, roomID, creatorUserID uint, access int, expireTime *time.Time) (*RoomInviteCode, error)
	JoinByInviteCode(ctx context.Context, userID uint, inviteCode string) (*RoomMember, error)
	ListInviteCodes(ctx context.Context, roomID, requesterUserID uint) ([]RoomInviteCode, error)
}
