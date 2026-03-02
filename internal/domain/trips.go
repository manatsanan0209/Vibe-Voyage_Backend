package domain

import (
	"context"
	"time"

	"gorm.io/gorm"
)

type Trips struct {
	TripID              uint           `json:"trip_id" gorm:"primaryKey;autoIncrement"`
	RoomID              uint           `json:"room_id" gorm:"not null;index"`
	Room                Room           `json:"-" gorm:"foreignKey:RoomID;references:RoomID;"`
	DestinationName     string         `json:"destination_name" gorm:"not null"`
	DestinationID       string         `json:"destination_id" gorm:"not null"`
	StartDate           time.Time      `json:"start_date" gorm:"not null"`
	EndDate             time.Time      `json:"end_date" gorm:"not null"`
	StructuredLifeStyle string         `json:"group_structured_lifestyle"`
	CreatedAt           time.Time      `json:"created_at"`
	UpdatedAt           time.Time      `json:"updated_at"`
	DeletedAt           gorm.DeletedAt `json:"deleted_at" gorm:"index"`
}

type PreferredDestination struct {
	DestinationName string `json:"destination_name"`
	DestinationID   string `json:"destination_id"`
}

type CreateTripInput struct {
	RoomName              string
	RoomImage             string
	DestinationName       string
	DestinationID         string
	StartDate             time.Time
	EndDate               time.Time
	PreferredDestinations []PreferredDestination
	TravelVibes           []string
	VoyagePriorities      []string
	FoodVibes             []string
	AdditionalNotes       string
}

type CreateTripResult struct {
	Room        *Room
	Trip        *Trips
	Member      *RoomMember
	Lifestyle   *UserLifestyle
	Suggestions []TripSchedule
}

type DaySchedule struct {
	DayNumber int
	Items     []TripSchedule
}

type GetTripScheduleResult struct {
	Trip        *Trips
	Suggestions []TripSchedule
	Days        []DaySchedule
}

type CreateTripScheduleInput struct {
	TripID        uint
	DayNumber     int
	SequenceOrder int
	PlaceName     string
	PlaceID       string
	Latitude      float64
	Longitude     float64
	StartTime     string
	EndTime       string
	Type          string
}

type RecommendedPlace struct {
	Name      string  `json:"name"`
	Category  string  `json:"category"`
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
}

type TripService interface {
	CreateTrip(ctx context.Context, userID uint, input CreateTripInput) (*CreateTripResult, error)
	GetTripSchedule(ctx context.Context, tripID uint) (*GetTripScheduleResult, error)
	CreateTripSchedule(ctx context.Context, inputs []CreateTripScheduleInput) ([]TripSchedule, error)
}
