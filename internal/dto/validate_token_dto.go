package dto

import "time"

type ValidateTokenRequestDTO struct {
	Token string `json:"token"`
}

type ValidateTokenResponseDTO struct {
	UserID    uint      `json:"user_id"`
	ExpiresAt time.Time `json:"expires_at"`
}
