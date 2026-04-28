package trip_suggestion

import (
	"github.com/gofiber/fiber/v2"
	"github.com/manatsanan0209/Vibe-Voyage_Backend/internal/domain"
	"github.com/manatsanan0209/Vibe-Voyage_Backend/internal/trip_suggestion/handler"
)

func Setup(app *fiber.App, svc domain.TripSuggestionService) {
	handler.NewTripSuggestionHandler(svc).RegisterRoutes(app)
}
