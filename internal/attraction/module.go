package attraction

import (
	"github.com/gofiber/fiber/v2"
	"github.com/manatsanan0209/Vibe-Voyage_Backend/internal/attraction/handler"
	attractionRepo "github.com/manatsanan0209/Vibe-Voyage_Backend/internal/attraction/repository"
	"github.com/manatsanan0209/Vibe-Voyage_Backend/internal/attraction/service"
	"gorm.io/gorm"
)

func Setup(app *fiber.App, db *gorm.DB) {
	repo := attractionRepo.NewAttractionRepository(db)
	svc := service.NewAttractionService(repo)
	handler.NewAttractionHandler(svc).RegisterRoutes(app)
}
