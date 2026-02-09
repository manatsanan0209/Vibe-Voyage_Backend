package server

import (
	"log"

	"github.com/gofiber/fiber/v2"
	"github.com/joho/godotenv"
	"github.com/manatsanan0209/Vibe-Voyage_Backend/internal/db"
	"github.com/manatsanan0209/Vibe-Voyage_Backend/internal/domain"

	userRepo "github.com/manatsanan0209/Vibe-Voyage_Backend/internal/user/repository"
	userService "github.com/manatsanan0209/Vibe-Voyage_Backend/internal/user/service"

	authPkg "github.com/manatsanan0209/Vibe-Voyage_Backend/internal/auth"
	healthPkg "github.com/manatsanan0209/Vibe-Voyage_Backend/internal/health"
	userPkg "github.com/manatsanan0209/Vibe-Voyage_Backend/internal/user"
)

func Run() error {
	_ = godotenv.Load()

	gormDB, err := db.Connect()
	if err != nil {
		return err
	}

	err = gormDB.AutoMigrate(&domain.User{})
	if err != nil {
		log.Fatal("Migration failed:", err)
	}
	log.Println("Database Migration Completed!")

	app := fiber.New()

	repo := userRepo.NewUserRepository(gormDB)
	svc := userService.NewUserService(repo)

	healthPkg.RegisterRoutes(app)
	userPkg.Setup(app, svc)
	authPkg.Setup(app, repo)

	return app.Listen(":8080")
}
