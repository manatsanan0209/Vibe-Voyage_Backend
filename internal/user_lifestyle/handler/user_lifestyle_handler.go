package handler

import (
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/manatsanan0209/Vibe-Voyage_Backend/internal/domain"
	"github.com/manatsanan0209/Vibe-Voyage_Backend/internal/dto"
)

type userLifestyleHandler struct {
	svc domain.UserLifestyleService
}

func NewUserLifestyleHandler(svc domain.UserLifestyleService) *userLifestyleHandler {
	return &userLifestyleHandler{svc: svc}
}

func (h *userLifestyleHandler) RegisterRoutes(app *fiber.App) {
	api := app.Group("/api/user_lifestyle")
	api.Post("/analyzelifestyle/:id", h.AnalyzeLifestyle)
}

func (h *userLifestyleHandler) AnalyzeLifestyle(c *fiber.Ctx) error {
	// authHeader := c.Get("Authorization")
	// if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
	// 	return c.Status(401).JSON(dto.APIResponse[any]{
	// 		Status:  401,
	// 		Message: "unauthorized",
	// 		Error:   "missing or invalid authorization header",
	// 	})
	// }

	// tokenStr := strings.TrimPrefix(authHeader, "Bearer ")
	// secret := os.Getenv("AUTH_TOKEN_SECRET")
	// if _, err := token.Validate(tokenStr, secret); err != nil {
	// 	return c.Status(401).JSON(dto.APIResponse[any]{
	// 		Status:  401,
	// 		Message: "unauthorized",
	// 		Error:   err.Error(),
	// 	})
	// }

	lifestyleID, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return c.Status(400).JSON(dto.APIResponse[any]{
			Status:  400,
			Message: "bad request",
			Error:   "id must be a number",
		})
	}

	places, err := h.svc.AnalyzeLifestyle(c.Context(), uint(lifestyleID))
	if err != nil {
		if err.Error() == "lifestyle not found" {
			return c.Status(404).JSON(dto.APIResponse[any]{
				Status:  404,
				Message: "not found",
				Error:   err.Error(),
			})
		}
		return c.Status(500).JSON(dto.APIResponse[any]{
			Status:  500,
			Message: "failed to analyze lifestyle",
			Error:   err.Error(),
		})
	}

	return c.Status(200).JSON(dto.APIResponse[[]domain.RecommendedPlace]{
		Status:  200,
		Message: "lifestyle analyzed successfully",
		Data:    &places,
	})
}
