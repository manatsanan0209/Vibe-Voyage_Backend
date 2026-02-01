package server

import (
	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Handler struct {
	pool *pgxpool.Pool
}

func NewHandler(pool *pgxpool.Pool) *Handler {
	return &Handler{pool: pool}
}

func (h *Handler) Register(app *fiber.App) {
	app.Get("/health", h.health)
}

func (h *Handler) health(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{"status": "ok"})
}
