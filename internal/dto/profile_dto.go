package dto

type ProfileResponseDTO struct {
	UserID       uint   `json:"user_id"`
	Username     string `json:"username"`
	FullName     string `json:"full_name"`
	Email        string `json:"email"`
	ProfileImage string `json:"profile_image"`
}

type UpdateProfileRequestDTO struct {
	Username     *string `json:"username"`
	FullName     *string `json:"full_name"`
	ProfileImage *string `json:"profile_image"`
}

type MyPostsResponseDTO struct {
	Total int64                     `json:"total"`
	Page  int                       `json:"page"`
	Limit int                       `json:"limit"`
	Posts []PublishedTripSummaryDTO `json:"posts"`
}
