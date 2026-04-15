package room

import (
	"github.com/gofiber/fiber/v2"
	"github.com/manatsanan0209/Vibe-Voyage_Backend/internal/domain"
	"github.com/manatsanan0209/Vibe-Voyage_Backend/internal/room/handler"
)

func Setup(app *fiber.App, svc domain.RoomService) {
	handler.NewRoomHandler(svc).RegisterRoutes(app)
}
