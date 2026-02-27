package hotel

import (
	"github.com/gofiber/fiber/v2"
	"github.com/manatsanan0209/Vibe-Voyage_Backend/internal/hotel/handler"
	hotelRepo "github.com/manatsanan0209/Vibe-Voyage_Backend/internal/hotel/repository"
	"github.com/manatsanan0209/Vibe-Voyage_Backend/internal/hotel/service"
	"gorm.io/gorm"
)

func Setup(app *fiber.App, db *gorm.DB) {
	repo := hotelRepo.NewHotelRepository(db)
	svc := service.NewHotelService(repo)
	handler.NewHotelHandler(svc).RegisterRoutes(app)
}
