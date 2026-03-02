package user_lifestyle

import (
	"github.com/gofiber/fiber/v2"
	"github.com/manatsanan0209/Vibe-Voyage_Backend/internal/user_lifestyle/handler"
	"github.com/manatsanan0209/Vibe-Voyage_Backend/internal/user_lifestyle/repository"
	"github.com/manatsanan0209/Vibe-Voyage_Backend/internal/user_lifestyle/service"
	"gorm.io/gorm"
)

func Setup(app *fiber.App, db *gorm.DB) {
	repo := repository.NewUserLifestyleRepository(db)
	svc := service.NewUserLifestyleService(repo, db)
	handler.NewUserLifestyleHandler(svc).RegisterRoutes(app)
}
