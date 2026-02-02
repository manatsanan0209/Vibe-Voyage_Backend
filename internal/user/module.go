package user

import (
    "github.com/gofiber/fiber/v2"
    "github.com/jackc/pgx/v5/pgxpool"
    "github.com/manatsanan0209/Vibe-Voyage_Backend/internal/user/handler"
    "github.com/manatsanan0209/Vibe-Voyage_Backend/internal/user/repository"
    "github.com/manatsanan0209/Vibe-Voyage_Backend/internal/user/service"
)

// Setup ทำหน้าที่ Wire Dependency ของ User ทั้งหมด
func Setup(app *fiber.App, pool *pgxpool.Pool) {
    repo := repository.NewUserRepository(pool)
    svc := service.NewUserService(repo)
    h := handler.NewUserHandler(svc)
    
    h.RegisterRoutes(app)
}