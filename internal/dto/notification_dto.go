package dto

import "time"

type NotificationResponseDTO struct {
	NotificationID uint      `json:"notification_id"`
	Type           string    `json:"type"`
	Title          string    `json:"title"`
	Message        string    `json:"message"`
	IsRead         bool      `json:"is_read"`
	ReferenceID    *uint     `json:"reference_id,omitempty"`
	ReferenceType  *string   `json:"reference_type,omitempty"`
	CreatedAt      time.Time `json:"created_at"`
}
