package domain

import (
	"context"
	"time"

	"gorm.io/gorm"
)

type TripSchedule struct {
	TripScheduleID uint           `json:"trip_schedule_id" gorm:"primaryKey;autoIncrement"`
	TripID         uint           `json:"trip_id" gorm:"not null;index"`
	Trip           Trips          `json:"-" gorm:"foreignKey:TripID;references:TripID"`
	DayNumber      int            `json:"day_number" gorm:"not null"`
	SequenceOrder  int            `json:"sequence_order" gorm:"not null"`
	PlaceName      string         `json:"place_name" gorm:"not null"`
	PlaceID        string         `json:"place_id" gorm:"not null"`
	Latitude       float64        `json:"latitude"`
	Longitude      float64        `json:"longitude"`
	StartTime      time.Time      `json:"start_time"`
	EndTime        time.Time      `json:"end_time"`
	Type           string         `json:"type" gorm:"not null"`
	CreatedAt      time.Time      `json:"created_at"`
	UpdatedAt      time.Time      `json:"updated_at"`
	DeletedAt      gorm.DeletedAt `json:"deleted_at" gorm:"index"`
}

type TripScheduleRepository interface {
	Create(ctx context.Context, schedule *TripSchedule) error
	GetByTripID(ctx context.Context, tripID uint) ([]TripSchedule, error)
	Update(ctx context.Context, schedule *TripSchedule) error
	Delete(ctx context.Context, scheduleID uint) error
}
