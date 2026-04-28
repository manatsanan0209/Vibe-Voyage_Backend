package dto

type PreferredDestinationDTO struct {
	DestinationName string  `json:"destination_name"`
	DestinationID   string  `json:"destination_id"`
	Latitude        float64 `json:"latitude"`
	Longitude       float64 `json:"longitude"`
}

type CreateTripRequestDTO struct {
	RoomName              string                    `json:"room_name"`
	RoomImage             string                    `json:"room_image"`
	DestinationName       string                    `json:"destination_name"`
	DestinationID         string                    `json:"destination_id"`
	StartDate             string                    `json:"start_date"`
	EndDate               string                    `json:"end_date"`
	PreferredDestinations []PreferredDestinationDTO `json:"preferred_destinations"`
	TravelVibes           []string                  `json:"travel_vibes"`
	VoyagePriorities      []string                  `json:"voyage_priorities"`
	FoodVibes             []string                  `json:"food_vibes"`
	AdditionalNotes       string                    `json:"additional_notes"`
}

type TripScheduleItemDTO struct {
	TripScheduleID uint    `json:"trip_schedule_id"`
	DayNumber      int     `json:"day_number"`
	SequenceOrder  int     `json:"sequence_order"`
	PlaceName      string  `json:"place_name"`
	PlaceID        string  `json:"place_id"`
	Latitude       float64 `json:"latitude"`
	Longitude      float64 `json:"longitude"`
	StartTime      string  `json:"start_time"`
	EndTime        string  `json:"end_time"`
	Type           string  `json:"type"`
}

type DayScheduleDTO struct {
	DayNumber int                   `json:"day_number"`
	Items     []TripScheduleItemDTO `json:"items"`
}

type GetTripScheduleResponseDTO struct {
	TripID          uint                  `json:"trip_id"`
	DestinationName string                `json:"destination_name"`
	StartDate       string                `json:"start_date"`
	EndDate         string                `json:"end_date"`
	IsPublished     bool                  `json:"is_published"`
	PublishedTripID *uint                 `json:"published_trip_id,omitempty"`
	Suggestions     []TripScheduleItemDTO `json:"suggestions"`
	Days            []DayScheduleDTO      `json:"days"`
}

type CreateTripScheduleItemRequestDTO struct {
	TripScheduleID uint    `json:"trip_schedule_id"`
	DayNumber      int     `json:"day_number"`
	SequenceOrder  int     `json:"sequence_order"`
	PlaceName      string  `json:"place_name"`
	PlaceID        string  `json:"place_id"`
	Latitude       float64 `json:"latitude"`
	Longitude      float64 `json:"longitude"`
	StartTime      string  `json:"start_time"`
	EndTime        string  `json:"end_time"`
	Type           string  `json:"type"`
}

type CreateTripScheduleRequestDTO struct {
	Items []CreateTripScheduleItemRequestDTO `json:"items"`
}

type RescheduleTripMemberScoreDTO struct {
	UserID         uint    `json:"user_id"`
	Username       string  `json:"username"`
	Score          float64 `json:"score"`
	EffectiveScore float64 `json:"effective_score"`
	TimesServed    int     `json:"times_served"`
	DeferredCount  int     `json:"deferred_count"`
}

type RescheduleTripResponseDTO struct {
	TripID           uint                           `json:"trip_id"`
	ScheduledCount   int                            `json:"scheduled_count"`
	SuggestionsCount int                            `json:"suggestions_count"`
	RoundCount       int                            `json:"round_count"`
	SelectedPlaceIDs []string                       `json:"selected_place_ids"`
	Scoreboard       []RescheduleTripMemberScoreDTO `json:"scoreboard"`
}

type RescheduleNotReadyMemberDTO struct {
	UserID      uint   `json:"user_id"`
	Username    string `json:"username"`
	LifestyleID *uint  `json:"lifestyle_id,omitempty"`
}

type RescheduleConflictResponseDTO struct {
	NotReadyMembers []RescheduleNotReadyMemberDTO `json:"not_ready_members"`
}

type CreateTripResponseDTO struct {
	RoomID          uint                  `json:"room_id"`
	TripID          uint                  `json:"trip_id"`
	LifestyleID     uint                  `json:"lifestyle_id"`
	RoomName        string                `json:"room_name"`
	RoomImage       string                `json:"room_image"`
	DestinationName string                `json:"destination_name"`
	StartDate       string                `json:"start_date"`
	EndDate         string                `json:"end_date"`
	Suggestions     []TripScheduleItemDTO `json:"suggestions"`
}

type JoinTripByInviteCodeRequestDTO struct {
	InviteCode string `json:"invite_code"`
}

type JoinTripByInviteCodeResponseDTO struct {
	TripID          uint   `json:"trip_id"`
	RoomID          uint   `json:"room_id"`
	DestinationName string `json:"destination_name"`
	StartDate       string `json:"start_date"`
	EndDate         string `json:"end_date"`
	RoomMemberID    uint   `json:"room_member_id"`
	UserID          uint   `json:"user_id"`
	Username        string `json:"username"`
	Role            int    `json:"role"`
	RoleName        string `json:"role_name"`
	JoinedAt        string `json:"joined_at"`
}
