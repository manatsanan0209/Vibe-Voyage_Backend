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
	TripScheduleID    uint    `json:"trip_schedule_id"`
	DayNumber         int     `json:"day_number"`
	SequenceOrder     int     `json:"sequence_order"`
	PlaceName         string  `json:"place_name"`
	PlaceID           string  `json:"place_id"`
	Latitude          float64 `json:"latitude"`
	Longitude         float64 `json:"longitude"`
	StartTime         string  `json:"start_time"`
	EndTime           string  `json:"end_time"`
	Type              string  `json:"type"`
	PlaceDetailStatus string  `json:"place_detail_status,omitempty"`
	PlaceDetail       any     `json:"place_detail,omitempty"`
}

type PlaceDetailOpeningHoursDTO struct {
	WeekdayText []string `json:"weekday_text"`
	OpenNow     bool     `json:"open_now"`
}

type PlaceDetailDTO struct {
	Rating           *float64                    `json:"rating,omitempty"`
	UserRatingCount  *int                        `json:"user_rating_count,omitempty"`
	OpeningHours     *PlaceDetailOpeningHoursDTO `json:"opening_hours,omitempty"`
	PhotoURL         string                      `json:"photo_url,omitempty"`
	GoogleMapsURI    string                      `json:"google_maps_uri,omitempty"`
	EditorialSummary string                      `json:"editorial_summary,omitempty"`
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

type PlanTripCurrentUserDTO struct {
	UserID        uint   `json:"user_id"`
	RoomMemberID  uint   `json:"room_member_id"`
	Role          int    `json:"role"`
	RoleName      string `json:"role_name"`
	CanEdit       bool   `json:"can_edit"`
	CanManageRoom bool   `json:"can_manage_room"`
}

type PlanTripMemberDTO struct {
	RoomMemberID          uint    `json:"room_member_id"`
	RoomID                uint    `json:"room_id"`
	UserID                uint    `json:"user_id"`
	Username              string  `json:"username"`
	ProfileImage          *string `json:"profile_image"`
	Role                  int     `json:"role"`
	RoleName              string  `json:"role_name"`
	HasSubmittedLifestyle bool    `json:"has_submitted_lifestyle"`
	HasAnalyzedLifestyle  bool    `json:"has_analyzed_lifestyle"`
}

type PlanTripWaitingMemberDTO struct {
	RoomMemberID uint    `json:"room_member_id"`
	UserID       uint    `json:"user_id"`
	Username     string  `json:"username"`
	ProfileImage *string `json:"profile_image"`
	LifestyleID  *uint   `json:"lifestyle_id"`
}

type PlanTripRescheduleReadinessDTO struct {
	Status         string                     `json:"status"`
	WaitingMembers []PlanTripWaitingMemberDTO `json:"waiting_members"`
}

type PlanTripPollingDTO struct {
	SchedulePollAfterMS          int `json:"schedule_poll_after_ms"`
	ScheduleReadinessPollAfterMS int `json:"schedule_readiness_poll_after_ms"`
}

type PlanTripBootstrapResponseDTO struct {
	TripID              uint                           `json:"trip_id"`
	RoomID              uint                           `json:"room_id"`
	CurrentUser         PlanTripCurrentUserDTO         `json:"current_user"`
	Schedule            GetTripScheduleResponseDTO     `json:"schedule"`
	Members             []PlanTripMemberDTO            `json:"members"`
	RescheduleReadiness PlanTripRescheduleReadinessDTO `json:"reschedule_readiness"`
	PublishStatus       *PublishStatusResponseDTO      `json:"publish_status"`
	Polling             PlanTripPollingDTO             `json:"polling"`
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

type TripFairnessSummaryDTO struct {
	TripID          uint    `json:"trip_id"`
	DestinationName string  `json:"destination_name"`
	GeneratedAt     string  `json:"generated_at"`
	RoundCount      int     `json:"round_count"`
	TotalPlaces     int     `json:"total_places"`
	GiniCoefficient float64 `json:"gini_coefficient"`
	FairnessRatio   float64 `json:"fairness_ratio"`
	ScoreStdDev     float64 `json:"score_std_dev"`
}

type AggregatedFairnessReportDTO struct {
	TripCount        int                      `json:"trip_count"`
	AvgGini          float64                  `json:"avg_gini_coefficient"`
	AvgFairnessRatio float64                  `json:"avg_fairness_ratio"`
	AvgScoreStdDev   float64                  `json:"avg_score_std_dev"`
	AvgTotalPlaces   float64                  `json:"avg_total_places"`
	Trips            []TripFairnessSummaryDTO `json:"trips"`
}

type FairnessReportMemberDTO struct {
	UserID         uint    `json:"user_id"`
	Username       string  `json:"username"`
	TimesServed    int     `json:"times_served"`
	Score          float64 `json:"score"`
	EffectiveScore float64 `json:"effective_score"`
	DeferredCount  int     `json:"deferred_count"`
	ScheduleShare  float64 `json:"schedule_share"`
	DeferredRate   float64 `json:"deferred_rate"`
}

type FairnessReportDTO struct {
	GeneratedAt      string                    `json:"generated_at"`
	AlgorithmVersion string                    `json:"algorithm_version"`
	RoundCount       int                       `json:"round_count"`
	TotalPlaces      int                       `json:"total_places"`
	GiniCoefficient  float64                   `json:"gini_coefficient"`
	FairnessRatio    float64                   `json:"fairness_ratio"`
	ScoreStdDev      float64                   `json:"score_std_dev"`
	Members          []FairnessReportMemberDTO `json:"members"`
}

type PlanTracePlaceDTO struct {
	Name      string  `json:"name"`
	PlaceID   string  `json:"place_id"`
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
}

type NearestNeighborStepDTO struct {
	Step       int               `json:"step"`
	From       PlanTracePlaceDTO `json:"from"`
	To         PlanTracePlaceDTO `json:"to"`
	DistanceKm float64           `json:"distance_km"`
}

type ScheduledPlaceTraceDTO struct {
	Name          string  `json:"name"`
	PlaceID       string  `json:"place_id"`
	Latitude      float64 `json:"latitude"`
	Longitude     float64 `json:"longitude"`
	DayNumber     int     `json:"day_number"`
	SequenceOrder int     `json:"sequence_order"`
}

type MealSelectionDetailDTO struct {
	MealType      string            `json:"meal_type"`
	DayNumber     int               `json:"day_number"`
	SequenceOrder int               `json:"sequence_order"`
	AnchorPlace   PlanTracePlaceDTO `json:"anchor_place"`
	SelectedPlace PlanTracePlaceDTO `json:"selected_place"`
	DistanceKm    float64           `json:"distance_km"`
}

type PlanTraceResponseDTO struct {
	TripID          uint   `json:"trip_id"`
	DestinationName string `json:"destination_name"`
	StartDate       string `json:"start_date"`
	EndDate         string `json:"end_date"`
	TotalDays       int    `json:"total_days"`
	PlacesPerDay    int    `json:"places_per_day"`

	Step1AIRecommendations    []PlanTracePlaceDTO      `json:"step1_ai_recommendations"`
	Step2NearestNeighborSteps []NearestNeighborStepDTO `json:"step2_nearest_neighbor_ordering"`
	Step2OrderedPlaces        []PlanTracePlaceDTO      `json:"step2_ordered_places"`
	Step3ScheduledPlaces      []ScheduledPlaceTraceDTO `json:"step3_scheduled_places"`
	Step3UnscheduledPlaces    []PlanTracePlaceDTO      `json:"step3_unscheduled_places"`
	Step4MealSelections       []MealSelectionDetailDTO `json:"step4_meal_selections"`
}

type GeoPointTraceDTO struct {
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
}

type RescheduleCandidateTraceDTO struct {
	Name      string  `json:"name"`
	PlaceID   string  `json:"place_id"`
	Category  string  `json:"category"`
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
}

type RescheduleMemberCandidateTraceDTO struct {
	UserID       uint                          `json:"user_id"`
	Username     string                        `json:"username"`
	Candidates   []RescheduleCandidateTraceDTO `json:"candidates"`
	CategoryRank map[string]int                `json:"category_rank"`
}

type RescheduleScoreUpdateTraceDTO struct {
	UserID   uint    `json:"user_id"`
	Username string  `json:"username"`
	Gained   float64 `json:"gained"`
	Reason   string  `json:"reason"`
	OldScore float64 `json:"old_score"`
	NewScore float64 `json:"new_score"`
}

type RescheduleMemberStateTraceDTO struct {
	UserID         uint    `json:"user_id"`
	Username       string  `json:"username"`
	Score          float64 `json:"score"`
	EffectiveScore float64 `json:"effective_score"`
	TimesServed    int     `json:"times_served"`
	DeferredCount  int     `json:"deferred_count"`
}

type FairnessRoundTraceDTO struct {
	Round                  int                             `json:"round"`
	PickedMemberID         uint                            `json:"picked_member_id"`
	PickedMemberUsername   string                          `json:"picked_member_username"`
	EffectiveScoreBefore   float64                         `json:"effective_score_before"`
	IsDeferred             bool                            `json:"is_deferred"`
	DeferReason            string                          `json:"defer_reason,omitempty"`
	SelectedPlace          *RescheduleCandidateTraceDTO    `json:"selected_place,omitempty"`
	DistanceFromPrevKm     float64                         `json:"distance_from_prev_km,omitempty"`
	DistanceFromCentroidKm float64                         `json:"distance_from_centroid_km,omitempty"`
	ScoreUpdates           []RescheduleScoreUpdateTraceDTO `json:"score_updates,omitempty"`
	MemberStatesAfter      []RescheduleMemberStateTraceDTO `json:"member_states_after"`
}

type ReschedulePlanTraceResponseDTO struct {
	TripID          uint   `json:"trip_id"`
	DestinationName string `json:"destination_name"`
	StartDate       string `json:"start_date"`
	EndDate         string `json:"end_date"`
	TotalDays       int    `json:"total_days"`
	PlacesPerDay    int    `json:"places_per_day"`

	Step1MembersAndCandidates []RescheduleMemberCandidateTraceDTO `json:"step1_members_and_candidates"`
	Step1Centroid             *GeoPointTraceDTO                   `json:"step1_centroid"`

	Step2FairnessRounds        []FairnessRoundTraceDTO       `json:"step2_fairness_rounds"`
	Step2FairnessOrderedPlaces []RescheduleCandidateTraceDTO `json:"step2_fairness_ordered_places"`
	Step2TotalRounds           int                           `json:"step2_total_rounds"`

	Step3NearestNeighborSteps []NearestNeighborStepDTO `json:"step3_nearest_neighbor_ordering"`
	Step3OrderedPlaces        []PlanTracePlaceDTO      `json:"step3_ordered_places"`

	Step4ScheduledPlaces   []ScheduledPlaceTraceDTO `json:"step4_scheduled_places"`
	Step4UnscheduledPlaces []PlanTracePlaceDTO      `json:"step4_unscheduled_places"`

	Step5MealSelections []MealSelectionDetailDTO `json:"step5_meal_selections"`
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
