package handler

import (
	"github.com/gofiber/fiber/v2"
	"github.com/manatsanan0209/Vibe-Voyage_Backend/internal/domain"
	"github.com/manatsanan0209/Vibe-Voyage_Backend/internal/dto"
)

type authHandler struct {
	svc domain.AuthService
}

func NewAuthHandler(svc domain.AuthService) *authHandler {
	return &authHandler{svc: svc}
}

func (h *authHandler) RegisterRoutes(app *fiber.App) {
	api := app.Group("/api/auth")
	api.Post("/register", h.Register)
	api.Post("/login", h.Login)
	api.Post("/validate", h.Validate)
}

func (h *authHandler) Register(c *fiber.Ctx) error {
	req := new(dto.RegisterRequestDTO)
	if err := c.BodyParser(req); err != nil {
		return c.Status(400).JSON(dto.APIResponse[any]{
			Status:  400,
			Message: "bad request",
			Error:   "invalid request body",
		})
	}

	user := &domain.User{
		Username: req.Username,
		Email:    req.Email,
		Password: req.Password,
		FullName: req.FullName,
	}

	tokenResult, err := h.svc.Register(c.Context(), user)
	if err != nil {
		return c.Status(400).JSON(dto.APIResponse[any]{
			Status:  400,
			Message: "register failed",
			Error:   err.Error(),
		})
	}

	registerResponseDTO := dto.NewRegisterResponseDTO(user, tokenResult.Token, tokenResult.ExpiresAt)
	return c.Status(201).JSON(dto.APIResponse[dto.RegisterResponseDTO]{
		Status:  201,
		Message: "register success",
		Data:    &registerResponseDTO,
	})
}

func (h *authHandler) Login(c *fiber.Ctx) error {
	req := new(dto.LoginRequestDTO)
	if err := c.BodyParser(req); err != nil {
		return c.Status(400).JSON(dto.APIResponse[any]{
			Status:  400,
			Message: "bad request",
			Error:   "invalid request body",
		})
	}

	user, tokenResult, err := h.svc.Login(c.Context(), req.Username, req.Password)
	if err != nil {
		return c.Status(401).JSON(dto.APIResponse[any]{
			Status:  401,
			Message: "login failed",
			Error:   err.Error(),
		})
	}

	loginResponseDTO := dto.NewLoginResponseDTO(user, tokenResult.Token, tokenResult.ExpiresAt)
	return c.Status(200).JSON(dto.APIResponse[dto.LoginResponseDTO]{
		Status:  200,
		Message: "login success",
		Data:    &loginResponseDTO,
	})
}

func (h *authHandler) Validate(c *fiber.Ctx) error {
	req := new(dto.ValidateTokenRequestDTO)
	if err := c.BodyParser(req); err != nil {
		return c.Status(400).JSON(dto.APIResponse[any]{
			Status:  400,
			Message: "bad request",
			Error:   "invalid request body",
		})
	}

	claims, err := h.svc.ValidateToken(c.Context(), req.Token)
	if err != nil {
		return c.Status(401).JSON(dto.APIResponse[any]{
			Status:  401,
			Message: "invalid token",
			Error:   err.Error(),
		})
	}

	resp := dto.ValidateTokenResponseDTO{
		UserID:    claims.UserID,
		ExpiresAt: claims.ExpiresAt,
	}

	return c.Status(200).JSON(dto.APIResponse[dto.ValidateTokenResponseDTO]{
		Status:  200,
		Message: "token valid",
		Data:    &resp,
	})
}
