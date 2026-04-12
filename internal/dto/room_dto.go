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

type UserRoomSummaryResponseDTO struct {
	RoomID        uint   `json:"room_id"`
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
	Access     string `json:"access"`
	ExpireTime string `json:"expire_time"`
}

type JoinRoomByInviteCodeRequestDTO struct {
	InviteCode string `json:"invite_code"`
}

type RoomInviteCodeResponseDTO struct {
	RoomInviteID        uint   `json:"room_invite_id"`
	RoomID              uint   `json:"room_id"`
	InviteCodeCreatorID uint   `json:"invite_code_creator_id"`
	InviteCode          string `json:"invite_code"`
	Access              string `json:"access"`
	ExpireTime          string `json:"expire_time"`
	CreatedAt           string `json:"created_at"`
}
