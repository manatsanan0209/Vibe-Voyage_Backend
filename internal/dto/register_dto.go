package dto

import (
	"time"

	"github.com/manatsanan0209/Vibe-Voyage_Backend/internal/domain"
)

type RegisterRequestDTO struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"password"`
	FullName string `json:"full_name"`
}

type RegisterResponseDTO struct {
	ID        uint      `json:"id"`
	Username  string    `json:"username"`
	Email     string    `json:"email"`
	FullName  string    `json:"full_name"`
	Token     string    `json:"token"`
	ExpiresAt time.Time `json:"expires_at"`
}

func NewRegisterResponseDTO(user *domain.User, token string, expiresAt time.Time) RegisterResponseDTO {
	if user == nil {
		return RegisterResponseDTO{}
	}

	return RegisterResponseDTO{
		ID:        user.UserID,
		Username:  user.Username,
		Email:     user.Email,
		FullName:  user.FullName,
		Token:     token,
		ExpiresAt: expiresAt,
	}
}
