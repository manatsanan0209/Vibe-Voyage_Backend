package handler

import (
	"strconv"

	"github.com/gofiber/fiber/v2"
	authMiddleware "github.com/manatsanan0209/Vibe-Voyage_Backend/internal/auth/middleware"
	"github.com/manatsanan0209/Vibe-Voyage_Backend/internal/domain"
	"github.com/manatsanan0209/Vibe-Voyage_Backend/internal/dto"
)

func (h *roomHandler) AddLifestyle(c *fiber.Ctx) error {
	userID, ok := authMiddleware.GetUserID(c)
	if !ok {
		return c.Status(401).JSON(dto.APIResponse[any]{
			Status:  401,
			Message: "unauthorized",
			Error:   "invalid token claims",
		})
	}

	roomID, err := strconv.ParseUint(c.Params("roomID"), 10, 64)
	if err != nil {
		return c.Status(400).JSON(dto.APIResponse[any]{
			Status:  400,
			Message: "bad request",
			Error:   "roomID must be a number",
		})
	}

	req := new(dto.AddRoomLifestyleRequestDTO)
	if err := c.BodyParser(req); err != nil {
		return c.Status(400).JSON(dto.APIResponse[any]{
			Status:  400,
			Message: "bad request",
			Error:   "invalid request body",
		})
	}

	preferredDestinations := make([]domain.PreferredDestination, 0, len(req.PreferredDestinations))
	for _, destination := range req.PreferredDestinations {
		preferredDestinations = append(preferredDestinations, domain.PreferredDestination{
			DestinationName: destination.DestinationName,
			DestinationID:   destination.DestinationID,
			Latitude:        destination.Latitude,
			Longitude:       destination.Longitude,
		})
	}

	input := domain.CreateRoomLifestyleInput{
		PreferredDestinations: preferredDestinations,
		TravelVibes:           req.TravelVibes,
		VoyagePriorities:      req.VoyagePriorities,
		FoodVibes:             req.FoodVibes,
		AdditionalNotes:       req.AdditionalNotes,
	}

	lifestyle, err := h.svc.AddRoomLifestyle(c.Context(), uint(roomID), userID, input)
	if err != nil {
		return c.Status(400).JSON(dto.APIResponse[any]{
			Status:  400,
			Message: "failed to add room lifestyle",
			Error:   err.Error(),
		})
	}

	resp := dto.RoomLifestyleResponseDTO{
		LifestyleID:           lifestyle.LifestyleID,
		UserID:                lifestyle.UserID,
		RoomID:                lifestyle.RoomID,
		PreferredDestinations: req.PreferredDestinations,
		TravelVibes:           req.TravelVibes,
		VoyagePriorities:      req.VoyagePriorities,
		FoodVibes:             req.FoodVibes,
		AdditionalNotes:       req.AdditionalNotes,
		CreatedAt:             lifestyle.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}

	return c.Status(201).JSON(dto.APIResponse[dto.RoomLifestyleResponseDTO]{
		Status:  201,
		Message: "room lifestyle added successfully",
		Data:    &resp,
	})
}
