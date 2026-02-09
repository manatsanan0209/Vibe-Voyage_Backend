package dto

import (
	"time"

	"github.com/manatsanan0209/Vibe-Voyage_Backend/internal/domain"
)

type LoginRequestDTO struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type LoginResponseDTO struct {
	ID        uint      `json:"id"`
	Username  string    `json:"username"`
	Token     string    `json:"token"`
	ExpiresAt time.Time `json:"expires_at"`
}

func NewLoginResponseDTO(user *domain.User, token string, expiresAt time.Time) LoginResponseDTO {
	if user == nil {
		return LoginResponseDTO{}
	}

	return LoginResponseDTO{
		ID:        user.UserID,
		Username:  user.Username,
		Token:     token,
		ExpiresAt: expiresAt,
	}
}
