package settings

import (
	"github.com/gofiber/fiber/v2"
	"github.com/manatsanan0209/Vibe-Voyage_Backend/internal/domain"
	"github.com/manatsanan0209/Vibe-Voyage_Backend/internal/settings/handler"
	"github.com/manatsanan0209/Vibe-Voyage_Backend/internal/settings/service"
)

func SetupWithRepo(app *fiber.App, repo domain.UserSettingsRepository) {
	svc := service.NewUserSettingsService(repo)
	handler.NewSettingsHandler(svc).RegisterRoutes(app)
}
