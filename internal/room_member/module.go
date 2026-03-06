package roommember

import (
	"github.com/gofiber/fiber/v2"
	"github.com/manatsanan0209/Vibe-Voyage_Backend/internal/room_member/handler"
	"github.com/manatsanan0209/Vibe-Voyage_Backend/internal/room_member/repository"
	"github.com/manatsanan0209/Vibe-Voyage_Backend/internal/room_member/service"
	"gorm.io/gorm"
)

func Setup(app *fiber.App, db *gorm.DB) {
	repo := repository.NewRoomMemberRepository(db)
	svc := service.NewRoomMemberService(repo)
	handler.NewRoomMemberHandler(svc).RegisterRoutes(app)
}
