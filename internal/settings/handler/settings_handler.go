package handler

import (
	"github.com/gofiber/fiber/v2"
	authMiddleware "github.com/manatsanan0209/Vibe-Voyage_Backend/internal/auth/middleware"
	"github.com/manatsanan0209/Vibe-Voyage_Backend/internal/domain"
	"github.com/manatsanan0209/Vibe-Voyage_Backend/internal/dto"
)

type settingsHandler struct {
	svc domain.UserSettingsService
}

func NewSettingsHandler(svc domain.UserSettingsService) *settingsHandler {
	return &settingsHandler{svc: svc}
}

func (h *settingsHandler) RegisterRoutes(app *fiber.App) {
	api := app.Group("/api/settings", authMiddleware.Authorize())
	api.Get("/", h.GetSettings)
	api.Patch("/", h.UpdateSettings)
}

func (h *settingsHandler) GetSettings(c *fiber.Ctx) error {
	userID, ok := authMiddleware.GetUserID(c)
	if !ok {
		return c.Status(401).JSON(dto.APIResponse[any]{
			Status:  401,
			Message: "unauthorized",
			Error:   "invalid token claims",
		})
	}

	settings, err := h.svc.GetSettings(c.Context(), userID)
	if err != nil {
		return c.Status(500).JSON(dto.APIResponse[any]{
			Status:  500,
			Message: "failed to get settings",
			Error:   err.Error(),
		})
	}

	resp := dto.UserSettingsResponseDTO{
		SettingsID:              settings.SettingsID,
		Theme:                   settings.Theme,
		Language:                settings.Language,
		DateFormat:              settings.DateFormat,
		TimeFormat:              settings.TimeFormat,
		NotifyRoomInvite:        settings.NotifyRoomInvite,
		NotifyMemberJoined:      settings.NotifyMemberJoined,
		NotifyMemberLeft:        settings.NotifyMemberLeft,
		NotifyTripCreated:       settings.NotifyTripCreated,
		NotifyLifestyleAnalyzed: settings.NotifyLifestyleAnalyzed,
		NotifyScheduleUpdated:   settings.NotifyScheduleUpdated,
	}

	return c.Status(200).JSON(dto.APIResponse[dto.UserSettingsResponseDTO]{
		Status:  200,
		Message: "success",
		Data:    &resp,
	})
}

func (h *settingsHandler) UpdateSettings(c *fiber.Ctx) error {
	userID, ok := authMiddleware.GetUserID(c)
	if !ok {
		return c.Status(401).JSON(dto.APIResponse[any]{
			Status:  401,
			Message: "unauthorized",
			Error:   "invalid token claims",
		})
	}

	req := new(dto.UpdateUserSettingsRequestDTO)
	if err := c.BodyParser(req); err != nil {
		return c.Status(400).JSON(dto.APIResponse[any]{
			Status:  400,
			Message: "bad request",
			Error:   "invalid request body",
		})
	}

	input := domain.UpdateUserSettingsInput{
		Theme:                   req.Theme,
		Language:                req.Language,
		DateFormat:              req.DateFormat,
		TimeFormat:              req.TimeFormat,
		NotifyRoomInvite:        req.NotifyRoomInvite,
		NotifyMemberJoined:      req.NotifyMemberJoined,
		NotifyMemberLeft:        req.NotifyMemberLeft,
		NotifyTripCreated:       req.NotifyTripCreated,
		NotifyLifestyleAnalyzed: req.NotifyLifestyleAnalyzed,
		NotifyScheduleUpdated:   req.NotifyScheduleUpdated,
	}

	settings, err := h.svc.UpdateSettings(c.Context(), userID, input)
	if err != nil {
		return c.Status(400).JSON(dto.APIResponse[any]{
			Status:  400,
			Message: "failed to update settings",
			Error:   err.Error(),
		})
	}

	resp := dto.UserSettingsResponseDTO{
		SettingsID:              settings.SettingsID,
		Theme:                   settings.Theme,
		Language:                settings.Language,
		DateFormat:              settings.DateFormat,
		TimeFormat:              settings.TimeFormat,
		NotifyRoomInvite:        settings.NotifyRoomInvite,
		NotifyMemberJoined:      settings.NotifyMemberJoined,
		NotifyMemberLeft:        settings.NotifyMemberLeft,
		NotifyTripCreated:       settings.NotifyTripCreated,
		NotifyLifestyleAnalyzed: settings.NotifyLifestyleAnalyzed,
		NotifyScheduleUpdated:   settings.NotifyScheduleUpdated,
	}

	return c.Status(200).JSON(dto.APIResponse[dto.UserSettingsResponseDTO]{
		Status:  200,
		Message: "settings updated",
		Data:    &resp,
	})
}
