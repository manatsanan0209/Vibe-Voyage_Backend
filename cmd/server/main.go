package server

import (
	"context"

	"github.com/gofiber/fiber/v2"
	"github.com/joho/godotenv"
	"github.com/manatsanan0209/Vibe-Voyage_Backend/internal/db"

	healthPkg "github.com/manatsanan0209/Vibe-Voyage_Backend/internal/health"
	userPkg "github.com/manatsanan0209/Vibe-Voyage_Backend/internal/user"
)

func Run() error {
	_ = godotenv.Load()
	ctx := context.Background()
	pool, err := db.Connect(ctx)
	if err != nil {
		return err
	}
	defer pool.Close()

	app := fiber.New()

	healthPkg.RegisterRoutes(app)
	userPkg.Setup(app, pool)

	return app.Listen(":8080")
}
