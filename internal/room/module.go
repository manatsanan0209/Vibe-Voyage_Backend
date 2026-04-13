package room

import (
	"github.com/gofiber/fiber/v2"
	"github.com/manatsanan0209/Vibe-Voyage_Backend/internal/room/handler"
	"github.com/manatsanan0209/Vibe-Voyage_Backend/internal/room/repository"
	"github.com/manatsanan0209/Vibe-Voyage_Backend/internal/room/service"
	userLifestyleRepo "github.com/manatsanan0209/Vibe-Voyage_Backend/internal/user_lifestyle/repository"
	"gorm.io/gorm"
)

func Setup(app *fiber.App, db *gorm.DB) {
	repo := repository.NewRoomRepository(db)
	inviteRepo := repository.NewRoomInviteCodeRepository(db)
	lifestyleRepo := userLifestyleRepo.NewUserLifestyleRepository(db)
	svc := service.NewRoomService(repo, inviteRepo, lifestyleRepo)
	handler.NewRoomHandler(svc).RegisterRoutes(app)
}
