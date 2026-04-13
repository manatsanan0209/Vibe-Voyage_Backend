package trip

import (
	"github.com/gofiber/fiber/v2"
	"github.com/manatsanan0209/Vibe-Voyage_Backend/internal/domain"
	"github.com/manatsanan0209/Vibe-Voyage_Backend/internal/trip/handler"
)

func Setup(app *fiber.App, svc domain.TripService) {
	handler.NewTripHandler(svc).RegisterRoutes(app)
}
