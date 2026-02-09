package auth

import (
	"github.com/gofiber/fiber/v2"
	"github.com/manatsanan0209/Vibe-Voyage_Backend/internal/auth/handler"
	"github.com/manatsanan0209/Vibe-Voyage_Backend/internal/auth/service"
	"github.com/manatsanan0209/Vibe-Voyage_Backend/internal/domain"
)

// รับ UserRepository เข้ามา
func Setup(app *fiber.App, userRepo domain.UserRepository) {
	authSvc := service.NewAuthService(userRepo)
	h := handler.NewAuthHandler(authSvc)
	h.RegisterRoutes(app)
}
