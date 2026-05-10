package dto

type RoomMemberResponseDTO struct {
	RoomMemberID uint   `json:"room_member_id"`
	RoomID       uint   `json:"room_id"`
	UserID       uint   `json:"user_id"`
	Username     string `json:"username"`
	Role         int    `json:"role"`
	RoleName     string `json:"role_name"`
	CreatedAt    string `json:"created_at"`
}

type RoomMemberLifestyleSubmissionResponseDTO struct {
	RoomMemberID          uint   `json:"room_member_id"`
	RoomID                uint   `json:"room_id"`
	UserID                uint   `json:"user_id"`
	Username              string `json:"username"`
	Role                  int    `json:"role"`
	RoleName              string `json:"role_name"`
	HasSubmittedLifestyle bool   `json:"has_submitted_lifestyle"`
	HasAnalyzedLifestyle  bool   `json:"has_analyzed_lifestyle"`
	LifestyleID           *uint  `json:"lifestyle_id"`
}

type UserRoomSummaryResponseDTO struct {
	RoomID        uint   `json:"room_id"`
	TripID        uint   `json:"trip_id"`
	RoomName      string `json:"room_name"`
	RoomImage     string `json:"room_image"`
	OwnerID       uint   `json:"owner_id"`
	OwnerUsername string `json:"owner_username"`
	Role          int    `json:"role"`
	RoleName      string `json:"role_name"`
	JoinedAt      string `json:"joined_at"`
	MembersCount  int64  `json:"members_count"`
}

type AddRoomMemberRequestDTO struct {
	UserID uint `json:"user_id"`
}

type CreateRoomInviteCodeRequestDTO struct {
	Access     int     `json:"access"`
	ExpireTime *string `json:"expire_time"`
}

type JoinRoomByInviteCodeRequestDTO struct {
	InviteCode string `json:"invite_code"`
}

type AddRoomLifestyleRequestDTO struct {
	PreferredDestinations []PreferredDestinationDTO `json:"preferred_destinations"`
	TravelVibes           []string                  `json:"travel_vibes"`
	VoyagePriorities      []string                  `json:"voyage_priorities"`
	FoodVibes             []string                  `json:"food_vibes"`
	AdditionalNotes       string                    `json:"additional_notes"`
}

type RoomLifestyleResponseDTO struct {
	LifestyleID           uint                      `json:"lifestyle_id"`
	UserID                uint                      `json:"user_id"`
	RoomID                uint                      `json:"room_id"`
	PreferredDestinations []PreferredDestinationDTO `json:"preferred_destinations"`
	TravelVibes           []string                  `json:"travel_vibes"`
	VoyagePriorities      []string                  `json:"voyage_priorities"`
	FoodVibes             []string                  `json:"food_vibes"`
	AdditionalNotes       string                    `json:"additional_notes"`
	CreatedAt             string                    `json:"created_at"`
}

type RoomInviteCodeResponseDTO struct {
	RoomInviteID        uint    `json:"room_invite_id"`
	RoomID              uint    `json:"room_id"`
	InviteCodeCreatorID uint    `json:"invite_code_creator_id"`
	InviteCode          string  `json:"invite_code"`
	Access              int     `json:"access"`
	AccessName          string  `json:"access_name"`
	ExpireTime          *string `json:"expire_time"`
	CreatedAt           string  `json:"created_at"`
}

type UpdateRoomRequestDTO struct {
	RoomName  *string `json:"room_name"`
	RoomImage *string `json:"room_image"`
}

type UpdateRoomResponseDTO struct {
	RoomID    uint   `json:"room_id"`
	OwnerID   uint   `json:"owner_id"`
	RoomName  string `json:"room_name"`
	RoomImage string `json:"room_image"`
	UpdatedAt string `json:"updated_at"`
}

type UpdateMemberRoleRequestDTO struct {
	Role int `json:"role"`
}

type TransferOwnershipRequestDTO struct {
	NewOwnerUserID uint `json:"new_owner_user_id"`
}
