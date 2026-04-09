package user_lifestyle

import (
	"github.com/gofiber/fiber/v2"
	"github.com/manatsanan0209/Vibe-Voyage_Backend/internal/domain"
	"github.com/manatsanan0209/Vibe-Voyage_Backend/internal/user_lifestyle/handler"
)

func Setup(app *fiber.App, svc domain.UserLifestyleService) {
	handler.NewUserLifestyleHandler(svc).RegisterRoutes(app)
}
