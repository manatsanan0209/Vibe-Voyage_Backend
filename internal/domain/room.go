package domain

import (
	"time"

	"gorm.io/gorm"
)

type Room struct {
	RoomID    uint           `json:"room_id" gorm:"primaryKey;autoIncrement"`
	OwnerID   uint           `json:"owner_id" gorm:"not null;index"`
	Owner     User           `json:"-" gorm:"foreignKey:OwnerID;references:UserID"`
	RoomName  string         `json:"room_name" gorm:"not null"`
	RoomImage string         `json:"room_image"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"deleted_at" gorm:"index"`
	Trips     []Trips        `json:"-" gorm:"foreignKey:RoomID;references:RoomID"`
}
