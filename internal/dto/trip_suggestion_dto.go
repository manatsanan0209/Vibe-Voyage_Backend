package dto

type PublishedTripPublisherDTO struct {
	UserID       uint   `json:"user_id"`
	Username     string `json:"username"`
	ProfileImage string `json:"profile_image"`
}

type PublishedTripSummaryDTO struct {
	PublishedTripID uint                      `json:"published_trip_id"`
	TripID          uint                      `json:"trip_id"`
	Title           string                    `json:"title"`
	Description     string                    `json:"description"`
	DestinationName string                    `json:"destination_name"`
	DestinationID   string                    `json:"destination_id"`
	StartDate       string                    `json:"start_date"`
	EndDate         string                    `json:"end_date"`
	ViewCount       int64                     `json:"view_count"`
	LikeCount       int64                     `json:"like_count"`
	Publisher       PublishedTripPublisherDTO  `json:"publisher"`
	IsLiked         bool                      `json:"is_liked"`
	IsBookmarked    bool                      `json:"is_bookmarked"`
	PublishedAt     string                    `json:"published_at"`
}

type PublishedTripDetailDTO struct {
	PublishedTripSummaryDTO
	ScheduleDays []DayScheduleDTO `json:"schedule_days"`
}

type GetTripFeedResponseDTO struct {
	Total int64                     `json:"total"`
	Page  int                       `json:"page"`
	Limit int                       `json:"limit"`
	Trips []PublishedTripSummaryDTO `json:"trips"`
}

type PublishTripRequestDTO struct {
	Title       string `json:"title"`
	Description string `json:"description"`
}

type PublishTripResponseDTO struct {
	PublishedTripID uint   `json:"published_trip_id"`
	TripID          uint   `json:"trip_id"`
	Title           string `json:"title"`
	Description     string `json:"description"`
	PublishedAt     string `json:"published_at"`
}

type ToggleLikeResponseDTO struct {
	Liked bool `json:"liked"`
}

type ToggleBookmarkResponseDTO struct {
	Bookmarked bool `json:"bookmarked"`
}

type UseAsTemplateRequestDTO struct {
	RoomName  string `json:"room_name"`
	RoomImage string `json:"room_image"`
	StartDate string `json:"start_date"`
	EndDate   string `json:"end_date"`
}

type PublishStatusResponseDTO struct {
	IsPublished     bool   `json:"is_published"`
	PublishedTripID *uint  `json:"published_trip_id,omitempty"`
	Title           string `json:"title,omitempty"`
	Description     string `json:"description,omitempty"`
	ViewCount       int64  `json:"view_count,omitempty"`
	LikeCount       int64  `json:"like_count,omitempty"`
	PublishedAt     string `json:"published_at,omitempty"`
}

type UseAsTemplateResponseDTO struct {
	RoomID          uint   `json:"room_id"`
	TripID          uint   `json:"trip_id"`
	RoomName        string `json:"room_name"`
	DestinationName string `json:"destination_name"`
	StartDate       string `json:"start_date"`
	EndDate         string `json:"end_date"`
}
