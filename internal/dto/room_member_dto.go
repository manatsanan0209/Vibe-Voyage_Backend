package dto

type RoomMemberDTO struct {
	RoomMemberID uint   `json:"room_member_id"`
	RoomID       uint   `json:"room_id"`
	UserID       uint   `json:"user_id"`
	Role         int    `json:"role"`
	RoleName     string `json:"role_name"`
	CreatedAt    string `json:"created_at"`
}

type AddRoomMemberRequestDTO struct {
	UserID uint `json:"user_id"`
}
