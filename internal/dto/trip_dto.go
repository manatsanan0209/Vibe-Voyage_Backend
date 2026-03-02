package dto

type PreferredDestinationDTO struct {
	DestinationName string `json:"destination_name"`
	DestinationID   string `json:"destination_id"`
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
	Suggestions []TripScheduleItemDTO `json:"suggestions"`
	Days        []DayScheduleDTO      `json:"days"`
}

type CreateTripScheduleItemRequestDTO struct {
	DayNumber     int     `json:"day_number"`
	SequenceOrder int     `json:"sequence_order"`
	PlaceName     string  `json:"place_name"`
	PlaceID       string  `json:"place_id"`
	Latitude      float64 `json:"latitude"`
	Longitude     float64 `json:"longitude"`
	StartTime     string  `json:"start_time"`
	EndTime       string  `json:"end_time"`
	Type          string  `json:"type"`
}

type CreateTripScheduleRequestDTO struct {
	Items []CreateTripScheduleItemRequestDTO `json:"items"`
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
