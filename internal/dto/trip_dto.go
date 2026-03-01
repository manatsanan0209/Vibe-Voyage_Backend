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
	VoyagePriorities      []string                  `json:"voyage_priorities"`
	FoodVibes             []string                  `json:"food_vibes"`
	AdditionalNotes       string                    `json:"additional_notes"`
}

type CreateTripResponseDTO struct {
	RoomID          uint   `json:"room_id"`
	TripID          uint   `json:"trip_id"`
	LifestyleID     uint   `json:"lifestyle_id"`
	RoomName        string `json:"room_name"`
	RoomImage       string `json:"room_image"`
	DestinationName string `json:"destination_name"`
	StartDate       string `json:"start_date"`
	EndDate         string `json:"end_date"`
}
