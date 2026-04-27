package dto

type UserSettingsResponseDTO struct {
	SettingsID              uint   `json:"settings_id"`
	Theme                   string `json:"theme"`
	Language                string `json:"language"`
	DateFormat              string `json:"date_format"`
	TimeFormat              string `json:"time_format"`
	NotifyRoomInvite        bool   `json:"notify_room_invite"`
	NotifyMemberJoined      bool   `json:"notify_member_joined"`
	NotifyMemberLeft        bool   `json:"notify_member_left"`
	NotifyTripCreated       bool   `json:"notify_trip_created"`
	NotifyLifestyleAnalyzed bool   `json:"notify_lifestyle_analyzed"`
	NotifyScheduleUpdated   bool   `json:"notify_schedule_updated"`
}

type UpdateUserSettingsRequestDTO struct {
	Theme                   *string `json:"theme"`
	Language                *string `json:"language"`
	DateFormat              *string `json:"date_format"`
	TimeFormat              *string `json:"time_format"`
	NotifyRoomInvite        *bool   `json:"notify_room_invite"`
	NotifyMemberJoined      *bool   `json:"notify_member_joined"`
	NotifyMemberLeft        *bool   `json:"notify_member_left"`
	NotifyTripCreated       *bool   `json:"notify_trip_created"`
	NotifyLifestyleAnalyzed *bool   `json:"notify_lifestyle_analyzed"`
	NotifyScheduleUpdated   *bool   `json:"notify_schedule_updated"`
}
