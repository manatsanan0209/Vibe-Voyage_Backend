package user

import (
	"github.com/gofiber/fiber/v2"
	"github.com/manatsanan0209/Vibe-Voyage_Backend/internal/domain"
	"github.com/manatsanan0209/Vibe-Voyage_Backend/internal/user/handler"
)

func Setup(app *fiber.App, svc domain.UserService, suggestionSvc domain.TripSuggestionService) {
	h := handler.NewUserHandler(svc, suggestionSvc)
	h.RegisterRoutes(app)
}
