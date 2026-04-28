package domain

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"gorm.io/gorm"
)

type Trips struct {
	TripID              uint           `json:"trip_id" gorm:"primaryKey;autoIncrement"`
	RoomID              uint           `json:"room_id" gorm:"not null;uniqueIndex"`
	Room                Room           `json:"-" gorm:"foreignKey:RoomID;references:RoomID;"`
	DestinationName     string         `json:"destination_name" gorm:"not null"`
	DestinationID       string         `json:"destination_id" gorm:"not null"`
	StartDate           time.Time      `json:"start_date" gorm:"not null"`
	EndDate             time.Time      `json:"end_date" gorm:"not null"`
	StructuredLifeStyle string         `json:"group_structured_lifestyle"`
	Version             int            `json:"version" gorm:"not null;default:0"`
	CreatedAt           time.Time      `json:"created_at"`
	UpdatedAt           time.Time      `json:"updated_at"`
	DeletedAt           gorm.DeletedAt `json:"deleted_at" gorm:"index"`
}

var (
	ErrForbidden                        = errors.New("forbidden")
	ErrLifestyleNotFound                = errors.New("lifestyle not found")
	ErrRescheduleConcurrentModification = errors.New("reschedule conflict: concurrent modification detected")
)

type PreferredDestination struct {
	DestinationName string  `json:"destination_name"`
	DestinationID   string  `json:"destination_id"`
	Latitude        float64 `json:"latitude"`
	Longitude       float64 `json:"longitude"`
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

type JoinTripByInviteCodeResult struct {
	Trip   *Trips
	Member *RoomMember
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

type RescheduleTripMemberScore struct {
	UserID         uint    `json:"user_id"`
	Username       string  `json:"username"`
	Score          float64 `json:"score"`
	EffectiveScore float64 `json:"effective_score"`
	TimesServed    int     `json:"times_served"`
	DeferredCount  int     `json:"deferred_count"`
}

type RescheduleTripResult struct {
	TripID           uint                        `json:"trip_id"`
	ScheduledCount   int                         `json:"scheduled_count"`
	SuggestionsCount int                         `json:"suggestions_count"`
	RoundCount       int                         `json:"round_count"`
	SelectedPlaceIDs []string                    `json:"selected_place_ids"`
	Scoreboard       []RescheduleTripMemberScore `json:"scoreboard"`
}

type RescheduleNotReadyMember struct {
	UserID      uint   `json:"user_id"`
	Username    string `json:"username"`
	LifestyleID *uint  `json:"lifestyle_id,omitempty"`
}

type RescheduleAnalysisNotReadyError struct {
	NotReadyMembers []RescheduleNotReadyMember
}

func (e *RescheduleAnalysisNotReadyError) Error() string {
	return "analysis_incomplete"
}

type CreateTripScheduleInput struct {
	TripScheduleID uint
	TripID         uint
	DayNumber      int
	SequenceOrder  int
	PlaceName      string
	PlaceID        string
	Latitude       float64
	Longitude      float64
	StartTime      string
	EndTime        string
	Type           string
}

type RecommendedPlace struct {
	PlaceID   string  `json:"place_id"`
	Name      string  `json:"name"`
	Category  string  `json:"category"`
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
}

type TripRepository interface {
	GetByID(ctx context.Context, tripID uint) (*Trips, error)
	GetByRoomID(ctx context.Context, roomID uint) (*Trips, error)
	IsUserInTripRoom(ctx context.Context, userID, tripID uint) (bool, error)
	GetUserRoleInTripRoom(ctx context.Context, userID, tripID uint) (int, bool, error)
	GetSchedulesByTripID(ctx context.Context, tripID uint) ([]TripSchedule, error)
	UpdateGroupStructuredLifestyle(ctx context.Context, tripID uint, snapshot string) error
	GetAttractionsByNames(ctx context.Context, names []string) (map[string][]Attraction, error)
	CreateTripBundle(
		ctx context.Context,
		userID uint,
		input CreateTripInput,
		preferredDestinationsJSON string,
		travelVibesJSON string,
		voyagePrioritiesJSON string,
		foodVibesJSON string,
	) (*CreateTripResult, error)
	CreateSchedules(ctx context.Context, schedules []TripSchedule) error
	ReplaceSchedulesByTripID(ctx context.Context, tripID uint, schedules []TripSchedule) error
	ReplaceScheduleAndSnapshot(ctx context.Context, tripID uint, schedules []TripSchedule, snapshot string) error
}

type TripService interface {
	CreateTrip(ctx context.Context, userID uint, input CreateTripInput) (*CreateTripResult, error)
	JoinTripByInviteCode(ctx context.Context, userID uint, inviteCode string) (*JoinTripByInviteCodeResult, error)
	GetTripSchedule(ctx context.Context, userID, tripID uint) (*GetTripScheduleResult, error)
	RescheduleTrip(ctx context.Context, userID, tripID uint) (*RescheduleTripResult, error)
	CreateTripSchedule(ctx context.Context, inputs []CreateTripScheduleInput) ([]TripSchedule, error)
	ReplaceTripSchedule(ctx context.Context, userID, tripID uint, inputs []CreateTripScheduleInput) ([]TripSchedule, error)
}

// IsStructuredLifestyleValid reports whether the raw JSON string is non-nil, valid JSON, and not trivially empty.
func IsStructuredLifestyleValid(value *string) bool {
	if value == nil {
		return false
	}

	var payload interface{}
	if json.Unmarshal([]byte(*value), &payload) != nil {
		return false
	}

	switch v := payload.(type) {
	case map[string]interface{}:
		return len(v) > 0
	case []interface{}:
		return len(v) > 0
	}

	return true
}
