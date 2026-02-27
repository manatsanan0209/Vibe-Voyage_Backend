package restaurant

import (
	"github.com/gofiber/fiber/v2"
	"github.com/manatsanan0209/Vibe-Voyage_Backend/internal/restaurant/handler"
	restaurantRepo "github.com/manatsanan0209/Vibe-Voyage_Backend/internal/restaurant/repository"
	"github.com/manatsanan0209/Vibe-Voyage_Backend/internal/restaurant/service"
	"gorm.io/gorm"
)

func Setup(app *fiber.App, db *gorm.DB) {
	repo := restaurantRepo.NewRestaurantRepository(db)
	svc := service.NewRestaurantService(repo)
	handler.NewRestaurantHandler(svc).RegisterRoutes(app)
}
