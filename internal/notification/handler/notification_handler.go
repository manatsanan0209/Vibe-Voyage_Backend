package handler

import (
	"strconv"

	"github.com/gofiber/fiber/v2"
	authMiddleware "github.com/manatsanan0209/Vibe-Voyage_Backend/internal/auth/middleware"
	"github.com/manatsanan0209/Vibe-Voyage_Backend/internal/domain"
	"github.com/manatsanan0209/Vibe-Voyage_Backend/internal/dto"
)

type notificationHandler struct {
	svc domain.NotificationService
}

func NewNotificationHandler(svc domain.NotificationService) *notificationHandler {
	return &notificationHandler{svc: svc}
}

func (h *notificationHandler) RegisterRoutes(app *fiber.App) {
	api := app.Group("/api/notifications", authMiddleware.Authorize())
	api.Get("/", h.GetNotifications)
	api.Patch("/read-all", h.MarkAllAsRead)
	api.Patch("/:id/read", h.MarkAsRead)
	api.Delete("/:id", h.Delete)
}

func (h *notificationHandler) GetNotifications(c *fiber.Ctx) error {
	userID, ok := authMiddleware.GetUserID(c)
	if !ok {
		return unauthorized(c)
	}

	unreadOnly := c.Query("unread") == "true"

	notifications, err := h.svc.GetNotifications(c.Context(), userID, unreadOnly)
	if err != nil {
		return c.Status(500).JSON(dto.APIResponse[any]{
			Status:  500,
			Message: "failed to get notifications",
			Error:   err.Error(),
		})
	}

	result := make([]dto.NotificationResponseDTO, 0, len(notifications))
	for _, n := range notifications {
		result = append(result, dto.NotificationResponseDTO{
			NotificationID: n.NotificationID,
			Type:           n.Type,
			Title:          n.Title,
			Message:        n.Message,
			IsRead:         n.IsRead,
			ReferenceID:    n.ReferenceID,
			ReferenceType:  n.ReferenceType,
			CreatedAt:      n.CreatedAt,
		})
	}

	return c.Status(200).JSON(dto.APIResponse[[]dto.NotificationResponseDTO]{
		Status:  200,
		Message: "success",
		Data:    &result,
	})
}

func (h *notificationHandler) MarkAsRead(c *fiber.Ctx) error {
	userID, ok := authMiddleware.GetUserID(c)
	if !ok {
		return unauthorized(c)
	}

	id, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return c.Status(400).JSON(dto.APIResponse[any]{
			Status:  400,
			Message: "bad request",
			Error:   "invalid notification id",
		})
	}

	if err := h.svc.MarkAsRead(c.Context(), uint(id), userID); err != nil {
		return c.Status(500).JSON(dto.APIResponse[any]{
			Status:  500,
			Message: "failed to mark as read",
			Error:   err.Error(),
		})
	}

	return c.Status(200).JSON(dto.APIResponse[any]{
		Status:  200,
		Message: "marked as read",
	})
}

func (h *notificationHandler) MarkAllAsRead(c *fiber.Ctx) error {
	userID, ok := authMiddleware.GetUserID(c)
	if !ok {
		return unauthorized(c)
	}

	if err := h.svc.MarkAllAsRead(c.Context(), userID); err != nil {
		return c.Status(500).JSON(dto.APIResponse[any]{
			Status:  500,
			Message: "failed to mark all as read",
			Error:   err.Error(),
		})
	}

	return c.Status(200).JSON(dto.APIResponse[any]{
		Status:  200,
		Message: "all notifications marked as read",
	})
}

func (h *notificationHandler) Delete(c *fiber.Ctx) error {
	userID, ok := authMiddleware.GetUserID(c)
	if !ok {
		return unauthorized(c)
	}

	id, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return c.Status(400).JSON(dto.APIResponse[any]{
			Status:  400,
			Message: "bad request",
			Error:   "invalid notification id",
		})
	}

	if err := h.svc.Delete(c.Context(), uint(id), userID); err != nil {
		return c.Status(500).JSON(dto.APIResponse[any]{
			Status:  500,
			Message: "failed to delete notification",
			Error:   err.Error(),
		})
	}

	return c.Status(200).JSON(dto.APIResponse[any]{
		Status:  200,
		Message: "notification deleted",
	})
}

func unauthorized(c *fiber.Ctx) error {
	return c.Status(401).JSON(dto.APIResponse[any]{
		Status:  401,
		Message: "unauthorized",
		Error:   "invalid token claims",
	})
}
