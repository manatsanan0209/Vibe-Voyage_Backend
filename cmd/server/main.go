package server

import (
	"context"

	"github.com/gofiber/fiber/v2"
	"github.com/manatsanan0209/Vibe-Voyage_Backend/internal/db"
)

func Run() error {
	ctx := context.Background()
	pool, err := db.Connect(ctx)
	if err != nil {
		return err
	}
	defer pool.Close()

	app := fiber.New()
	handler := NewHandler(pool)
	handler.Register(app)

	return app.Listen(":8080")
}
