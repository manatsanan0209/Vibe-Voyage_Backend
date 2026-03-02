package trip

import (
	"github.com/gofiber/fiber/v2"
	"github.com/manatsanan0209/Vibe-Voyage_Backend/internal/trip/handler"
	"github.com/manatsanan0209/Vibe-Voyage_Backend/internal/trip/service"
	userLifestyleRepo "github.com/manatsanan0209/Vibe-Voyage_Backend/internal/user_lifestyle/repository"
	userLifestyleSvc "github.com/manatsanan0209/Vibe-Voyage_Backend/internal/user_lifestyle/service"
	"gorm.io/gorm"
)

func Setup(app *fiber.App, db *gorm.DB) {
	lifestyleRepo := userLifestyleRepo.NewUserLifestyleRepository(db)
	lifestyleSvc := userLifestyleSvc.NewUserLifestyleService(lifestyleRepo, db)
	svc := service.NewTripService(db, lifestyleSvc)
	handler.NewTripHandler(svc).RegisterRoutes(app)
}
