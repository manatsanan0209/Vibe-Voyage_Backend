package handler

import (
	"github.com/gofiber/fiber/v2"
	"github.com/manatsanan0209/Vibe-Voyage_Backend/internal/domain"
)

type userHandler struct {
	svc domain.UserService
}

func NewUserHandler(svc domain.UserService) *userHandler {
	return &userHandler{svc: svc}
}

func (h *userHandler) RegisterRoutes(app *fiber.App) {
	api := app.Group("/api/users")
	api.Get("/:id", h.GetUser)
}


func (h *userHandler) GetUser(c *fiber.Ctx) error {
	// ... implementation ...
	return c.SendStatus(200)
}
