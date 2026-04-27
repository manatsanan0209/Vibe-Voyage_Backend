package notification

import (
	"github.com/gofiber/fiber/v2"
	"github.com/manatsanan0209/Vibe-Voyage_Backend/internal/domain"
	"github.com/manatsanan0209/Vibe-Voyage_Backend/internal/notification/handler"
	"github.com/manatsanan0209/Vibe-Voyage_Backend/internal/notification/service"
)

func SetupService(repo domain.NotificationRepository, settingsRepo domain.UserSettingsRepository) domain.NotificationService {
	return service.NewNotificationService(repo, settingsRepo)
}

func SetupHandler(app *fiber.App, svc domain.NotificationService) {
	handler.NewNotificationHandler(svc).RegisterRoutes(app)
}
