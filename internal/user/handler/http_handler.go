package handler

import (
	"github.com/gofiber/fiber/v2"
	"github.com/manatsanan0209/Vibe-Voyage_Backend/internal/domain"
)

type userHandler struct {
	svc domain.UserService // เรียกผ่าน Interface
}

func NewUserHandler(svc domain.UserService) *userHandler {
	return &userHandler{svc: svc}
}

func (h *userHandler) RegisterRoutes(app *fiber.App) {
	api := app.Group("/api/users")
	api.Post("/", h.register)
	api.Get("/:id", h.getUser)
}

func (h *userHandler) register(c *fiber.Ctx) error {
	user := new(domain.User)
	if err := c.BodyParser(user); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "bad request"})
	}

	if err := h.svc.Register(c.Context(), user); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	return c.Status(201).JSON(user)
}

func (h *userHandler) getUser(c *fiber.Ctx) error {
	// ... implementation ...
	return c.SendStatus(200)
}