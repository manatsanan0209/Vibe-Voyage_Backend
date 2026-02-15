package server

import (
	"log"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
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
	app.Use(cors.New(cors.Config{
		AllowOrigins: "*",
		AllowMethods: "GET,POST,PUT,PATCH,DELETE,OPTIONS",
		AllowHeaders: "Origin,Content-Type,Accept,Authorization",
	}))
	app.Use(logger.New(logger.Config{
		TimeFormat: "2006-01-02 15:04:05",
		Format:     "${time} | ${status} | ${latency} | ${method} ${path}\n",
	}))

	repo := userRepo.NewUserRepository(gormDB)
	svc := userService.NewUserService(repo)

	healthPkg.RegisterRoutes(app)
	userPkg.Setup(app, svc)
	authPkg.Setup(app, repo)

	return app.Listen(":8080")
}
