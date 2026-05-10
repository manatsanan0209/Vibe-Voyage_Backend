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
	StructuredLifeStyle string         `json:"group_structured_lifestyle" gorm:"column:structured_lifestyle"`
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
	Trip         *Trips
	Suggestions  []TripSchedule
	Days         []DaySchedule
	PlaceDetails map[string]PlaceDetailAttachment
}

type TripRoomMembership struct {
	Trip         *Trips
	RoomMemberID uint
	Role         int
}

type PlanTripPublishStatus struct {
	IsPublished     bool
	PublishedTripID *uint
	Title           string
	Description     string
	ViewCount       int64
	LikeCount       int64
	PublishedAt     *time.Time
}

type PlanTripBootstrapResult struct {
	Trip                 *Trips
	CurrentMember        TripRoomMembership
	Suggestions          []TripSchedule
	Days                 []DaySchedule
	PlaceDetails         map[string]PlaceDetailAttachment
	Members              []MemberLifestyleSubmissionStatus
	PublishStatus        *PlanTripPublishStatus
	SchedulePollAfterMS  int
	ReadinessPollAfterMS int
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

type TripFairnessSummary struct {
	TripID          uint    `json:"trip_id"`
	DestinationName string  `json:"destination_name"`
	GeneratedAt     string  `json:"generated_at"`
	RoundCount      int     `json:"round_count"`
	TotalPlaces     int     `json:"total_places"`
	GiniCoefficient float64 `json:"gini_coefficient"`
	FairnessRatio   float64 `json:"fairness_ratio"`
	ScoreStdDev     float64 `json:"score_std_dev"`
}

type AggregatedFairnessReport struct {
	TripCount        int                   `json:"trip_count"`
	AvgGini          float64               `json:"avg_gini_coefficient"`
	AvgFairnessRatio float64               `json:"avg_fairness_ratio"`
	AvgScoreStdDev   float64               `json:"avg_score_std_dev"`
	AvgTotalPlaces   float64               `json:"avg_total_places"`
	Trips            []TripFairnessSummary `json:"trips"`
}

type FairnessReportMember struct {
	UserID         uint    `json:"user_id"`
	Username       string  `json:"username"`
	TimesServed    int     `json:"times_served"`
	Score          float64 `json:"score"`
	EffectiveScore float64 `json:"effective_score"`
	DeferredCount  int     `json:"deferred_count"`
	ScheduleShare  float64 `json:"schedule_share"`
	DeferredRate   float64 `json:"deferred_rate"`
}

type FairnessReport struct {
	GeneratedAt      string                 `json:"generated_at"`
	AlgorithmVersion string                 `json:"algorithm_version"`
	RoundCount       int                    `json:"round_count"`
	TotalPlaces      int                    `json:"total_places"`
	GiniCoefficient  float64                `json:"gini_coefficient"`
	FairnessRatio    float64                `json:"fairness_ratio"`
	ScoreStdDev      float64                `json:"score_std_dev"`
	Members          []FairnessReportMember `json:"members"`
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
	GetTripRoomMembership(ctx context.Context, userID, tripID uint) (*TripRoomMembership, bool, error)
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
	GetTripsByUserID(ctx context.Context, userID uint) ([]Trips, error)
}

type TripService interface {
	CreateTrip(ctx context.Context, userID uint, input CreateTripInput) (*CreateTripResult, error)
	JoinTripByInviteCode(ctx context.Context, userID uint, inviteCode string) (*JoinTripByInviteCodeResult, error)
	GetTripSchedule(ctx context.Context, userID, tripID uint) (*GetTripScheduleResult, error)
	GetPlanTripBootstrap(ctx context.Context, userID, tripID uint) (*PlanTripBootstrapResult, error)
	RescheduleTrip(ctx context.Context, userID, tripID uint) (*RescheduleTripResult, error)
	GetFairnessReport(ctx context.Context, userID, tripID uint) (*FairnessReport, error)
	GetAggregatedFairnessReport(ctx context.Context, userID uint) (*AggregatedFairnessReport, error)
	CreateTripSchedule(ctx context.Context, inputs []CreateTripScheduleInput) ([]TripSchedule, error)
	ReplaceTripSchedule(ctx context.Context, userID, tripID uint, inputs []CreateTripScheduleInput) ([]TripSchedule, error)
	GetPlanTrace(ctx context.Context, userID, tripID uint) (*PlanTrace, error)
	GetReschedulePlanTrace(ctx context.Context, userID, tripID uint) (*ReschedulePlanTrace, error)
}

type PlanTracePlace struct {
	Name      string  `json:"name"`
	PlaceID   string  `json:"place_id"`
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
}

type NearestNeighborStep struct {
	Step       int            `json:"step"`
	From       PlanTracePlace `json:"from"`
	To         PlanTracePlace `json:"to"`
	DistanceKm float64        `json:"distance_km"`
}

type ScheduledPlaceTrace struct {
	Name          string  `json:"name"`
	PlaceID       string  `json:"place_id"`
	Latitude      float64 `json:"latitude"`
	Longitude     float64 `json:"longitude"`
	DayNumber     int     `json:"day_number"`
	SequenceOrder int     `json:"sequence_order"`
}

type MealSelectionDetail struct {
	MealType      string         `json:"meal_type"`
	DayNumber     int            `json:"day_number"`
	SequenceOrder int            `json:"sequence_order"`
	AnchorPlace   PlanTracePlace `json:"anchor_place"`
	SelectedPlace PlanTracePlace `json:"selected_place"`
	DistanceKm    float64        `json:"distance_km"`
}

type PlanTrace struct {
	TripID          uint   `json:"trip_id"`
	DestinationName string `json:"destination_name"`
	StartDate       string `json:"start_date"`
	EndDate         string `json:"end_date"`
	TotalDays       int    `json:"total_days"`
	PlacesPerDay    int    `json:"places_per_day"`

	AIRecommendations    []PlanTracePlace      `json:"step1_ai_recommendations"`
	NearestNeighborSteps []NearestNeighborStep `json:"step2_nearest_neighbor_ordering"`
	OrderedPlaces        []PlanTracePlace      `json:"step2_ordered_places"`
	ScheduledPlaces      []ScheduledPlaceTrace `json:"step3_scheduled_places"`
	UnscheduledPlaces    []PlanTracePlace      `json:"step3_unscheduled_places"`
	MealSelections       []MealSelectionDetail `json:"step4_meal_selections"`
}

type GeoPointTrace struct {
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
}

type RescheduleCandidateTrace struct {
	Name      string  `json:"name"`
	PlaceID   string  `json:"place_id"`
	Category  string  `json:"category"`
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
}

type RescheduleMemberCandidateTrace struct {
	UserID       uint                       `json:"user_id"`
	Username     string                     `json:"username"`
	Candidates   []RescheduleCandidateTrace `json:"candidates"`
	CategoryRank map[string]int             `json:"category_rank"`
}

type RescheduleScoreUpdateTrace struct {
	UserID   uint    `json:"user_id"`
	Username string  `json:"username"`
	Gained   float64 `json:"gained"`
	Reason   string  `json:"reason"`
	OldScore float64 `json:"old_score"`
	NewScore float64 `json:"new_score"`
}

type RescheduleMemberStateTrace struct {
	UserID         uint    `json:"user_id"`
	Username       string  `json:"username"`
	Score          float64 `json:"score"`
	EffectiveScore float64 `json:"effective_score"`
	TimesServed    int     `json:"times_served"`
	DeferredCount  int     `json:"deferred_count"`
}

type FairnessRoundTrace struct {
	Round                  int                          `json:"round"`
	PickedMemberID         uint                         `json:"picked_member_id"`
	PickedMemberUsername   string                       `json:"picked_member_username"`
	EffectiveScoreBefore   float64                      `json:"effective_score_before"`
	IsDeferred             bool                         `json:"is_deferred"`
	DeferReason            string                       `json:"defer_reason,omitempty"`
	SelectedPlace          *RescheduleCandidateTrace    `json:"selected_place,omitempty"`
	DistanceFromPrevKm     float64                      `json:"distance_from_prev_km,omitempty"`
	DistanceFromCentroidKm float64                      `json:"distance_from_centroid_km,omitempty"`
	ScoreUpdates           []RescheduleScoreUpdateTrace `json:"score_updates,omitempty"`
	MemberStatesAfter      []RescheduleMemberStateTrace `json:"member_states_after"`
}

type ReschedulePlanTrace struct {
	TripID          uint   `json:"trip_id"`
	DestinationName string `json:"destination_name"`
	StartDate       string `json:"start_date"`
	EndDate         string `json:"end_date"`
	TotalDays       int    `json:"total_days"`
	PlacesPerDay    int    `json:"places_per_day"`

	Step1Members  []RescheduleMemberCandidateTrace `json:"step1_members_and_candidates"`
	Step1Centroid *GeoPointTrace                   `json:"step1_centroid"`

	Step2FairnessRounds        []FairnessRoundTrace       `json:"step2_fairness_rounds"`
	Step2FairnessOrderedPlaces []RescheduleCandidateTrace `json:"step2_fairness_ordered_places"`
	Step2TotalRounds           int                        `json:"step2_total_rounds"`

	Step3NearestNeighborSteps []NearestNeighborStep `json:"step3_nearest_neighbor_ordering"`
	Step3OrderedPlaces        []PlanTracePlace      `json:"step3_ordered_places"`

	Step4ScheduledPlaces   []ScheduledPlaceTrace `json:"step4_scheduled_places"`
	Step4UnscheduledPlaces []PlanTracePlace      `json:"step4_unscheduled_places"`

	Step5MealSelections []MealSelectionDetail `json:"step5_meal_selections"`
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
