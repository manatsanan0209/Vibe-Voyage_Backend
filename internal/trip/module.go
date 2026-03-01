package trip

import (
	"github.com/gofiber/fiber/v2"
	"github.com/manatsanan0209/Vibe-Voyage_Backend/internal/trip/handler"
	"github.com/manatsanan0209/Vibe-Voyage_Backend/internal/trip/service"
	"gorm.io/gorm"
)

func Setup(app *fiber.App, db *gorm.DB) {
	svc := service.NewTripService(db)
	handler.NewTripHandler(svc).RegisterRoutes(app)
}
