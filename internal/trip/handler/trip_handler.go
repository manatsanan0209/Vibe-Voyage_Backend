package handler

import (
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/manatsanan0209/Vibe-Voyage_Backend/internal/auth/token"
	"github.com/manatsanan0209/Vibe-Voyage_Backend/internal/domain"
	"github.com/manatsanan0209/Vibe-Voyage_Backend/internal/dto"
)

type tripHandler struct {
	svc domain.TripService
}

func NewTripHandler(svc domain.TripService) *tripHandler {
	return &tripHandler{svc: svc}
}

func (h *tripHandler) RegisterRoutes(app *fiber.App) {
	api := app.Group("/api/trip")
	api.Post("/", h.CreateTrip)
	api.Get("/:tripID/schedule", h.GetTripSchedule)
	api.Post("/:tripID/schedule", h.CreateTripSchedule)
}

func toScheduleItemDTO(item domain.TripSchedule) dto.TripScheduleItemDTO {
	return dto.TripScheduleItemDTO{
		TripScheduleID: item.TripScheduleID,
		DayNumber:      item.DayNumber,
		SequenceOrder:  item.SequenceOrder,
		PlaceName:      item.PlaceName,
		PlaceID:        item.PlaceID,
		Latitude:       item.Latitude,
		Longitude:      item.Longitude,
		StartTime:      item.StartTime.Format("15:04"),
		EndTime:        item.EndTime.Format("15:04"),
		Type:           item.Type,
	}
}

func (h *tripHandler) GetTripSchedule(c *fiber.Ctx) error {
	authHeader := c.Get("Authorization")
	if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
		return c.Status(401).JSON(dto.APIResponse[any]{
			Status:  401,
			Message: "unauthorized",
			Error:   "missing or invalid authorization header",
		})
	}

	tokenStr := strings.TrimPrefix(authHeader, "Bearer ")
	secret := os.Getenv("AUTH_TOKEN_SECRET")
	if _, err := token.Validate(tokenStr, secret); err != nil {
		return c.Status(401).JSON(dto.APIResponse[any]{
			Status:  401,
			Message: "unauthorized",
			Error:   err.Error(),
		})
	}

	tripID, err := strconv.ParseUint(c.Params("tripID"), 10, 64)
	if err != nil {
		return c.Status(400).JSON(dto.APIResponse[any]{
			Status:  400,
			Message: "bad request",
			Error:   "tripID must be a number",
		})
	}

	result, err := h.svc.GetTripSchedule(c.Context(), uint(tripID))
	if err != nil {
		return c.Status(500).JSON(dto.APIResponse[any]{
			Status:  500,
			Message: "failed to get trip schedule",
			Error:   err.Error(),
		})
	}

	suggestions := make([]dto.TripScheduleItemDTO, 0, len(result.Suggestions))
	for _, item := range result.Suggestions {
		suggestions = append(suggestions, toScheduleItemDTO(item))
	}

	days := make([]dto.DayScheduleDTO, 0, len(result.Days))
	for _, day := range result.Days {
		items := make([]dto.TripScheduleItemDTO, 0, len(day.Items))
		for _, item := range day.Items {
			items = append(items, toScheduleItemDTO(item))
		}
		days = append(days, dto.DayScheduleDTO{
			DayNumber: day.DayNumber,
			Items:     items,
		})
	}

	resp := dto.GetTripScheduleResponseDTO{
		Suggestions: suggestions,
		Days:        days,
	}

	return c.Status(200).JSON(dto.APIResponse[dto.GetTripScheduleResponseDTO]{
		Status:  200,
		Message: "success",
		Data:    &resp,
	})
}

func (h *tripHandler) CreateTrip(c *fiber.Ctx) error {
	authHeader := c.Get("Authorization")
	if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
		return c.Status(401).JSON(dto.APIResponse[any]{
			Status:  401,
			Message: "unauthorized",
			Error:   "missing or invalid authorization header",
		})
	}

	tokenStr := strings.TrimPrefix(authHeader, "Bearer ")
	secret := os.Getenv("AUTH_TOKEN_SECRET")
	claims, err := token.Validate(tokenStr, secret)
	if err != nil {
		return c.Status(401).JSON(dto.APIResponse[any]{
			Status:  401,
			Message: "unauthorized",
			Error:   err.Error(),
		})
	}

	req := new(dto.CreateTripRequestDTO)
	if err := c.BodyParser(req); err != nil {
		return c.Status(400).JSON(dto.APIResponse[any]{
			Status:  400,
			Message: "bad request",
			Error:   "invalid request body",
		})
	}

	startDate, err := time.Parse("2006-01-02", req.StartDate)
	if err != nil {
		return c.Status(400).JSON(dto.APIResponse[any]{
			Status:  400,
			Message: "bad request",
			Error:   "start_date must be in YYYY-MM-DD format",
		})
	}

	endDate, err := time.Parse("2006-01-02", req.EndDate)
	if err != nil {
		return c.Status(400).JSON(dto.APIResponse[any]{
			Status:  400,
			Message: "bad request",
			Error:   "end_date must be in YYYY-MM-DD format",
		})
	}

	preferredDests := make([]domain.PreferredDestination, len(req.PreferredDestinations))
	for i, d := range req.PreferredDestinations {
		preferredDests[i] = domain.PreferredDestination{
			DestinationName: d.DestinationName,
			DestinationID:   d.DestinationID,
		}
	}

	input := domain.CreateTripInput{
		RoomName:              req.RoomName,
		RoomImage:             req.RoomImage,
		DestinationName:       req.DestinationName,
		DestinationID:         req.DestinationID,
		StartDate:             startDate,
		EndDate:               endDate,
		PreferredDestinations: preferredDests,
		VoyagePriorities:      req.VoyagePriorities,
		FoodVibes:             req.FoodVibes,
		AdditionalNotes:       req.AdditionalNotes,
	}

	result, err := h.svc.CreateTrip(c.Context(), claims.UserID, input)
	if err != nil {
		return c.Status(400).JSON(dto.APIResponse[any]{
			Status:  400,
			Message: "create trip failed",
			Error:   err.Error(),
		})
	}

	resp := dto.CreateTripResponseDTO{
		RoomID:          result.Room.RoomID,
		TripID:          result.Trip.TripID,
		LifestyleID:     result.Lifestyle.LifestyleID,
		RoomName:        result.Room.RoomName,
		RoomImage:       result.Room.RoomImage,
		DestinationName: result.Trip.DestinationName,
		StartDate:       result.Trip.StartDate.Format("2006-01-02"),
		EndDate:         result.Trip.EndDate.Format("2006-01-02"),
	}

	return c.Status(201).JSON(dto.APIResponse[dto.CreateTripResponseDTO]{
		Status:  201,
		Message: "trip created successfully",
		Data:    &resp,
	})
}

func (h *tripHandler) CreateTripSchedule(c *fiber.Ctx) error {
	authHeader := c.Get("Authorization")
	if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
		return c.Status(401).JSON(dto.APIResponse[any]{
			Status:  401,
			Message: "unauthorized",
			Error:   "missing or invalid authorization header",
		})
	}

	tokenStr := strings.TrimPrefix(authHeader, "Bearer ")
	secret := os.Getenv("AUTH_TOKEN_SECRET")
	if _, err := token.Validate(tokenStr, secret); err != nil {
		return c.Status(401).JSON(dto.APIResponse[any]{
			Status:  401,
			Message: "unauthorized",
			Error:   err.Error(),
		})
	}

	tripID, err := strconv.ParseUint(c.Params("tripID"), 10, 64)
	if err != nil {
		return c.Status(400).JSON(dto.APIResponse[any]{
			Status:  400,
			Message: "bad request",
			Error:   "tripID must be a number",
		})
	}

	req := new(dto.CreateTripScheduleRequestDTO)
	if err := c.BodyParser(req); err != nil {
		return c.Status(400).JSON(dto.APIResponse[any]{
			Status:  400,
			Message: "bad request",
			Error:   "invalid request body",
		})
	}

	inputs := make([]domain.CreateTripScheduleInput, len(req.Items))
	for i, item := range req.Items {
		inputs[i] = domain.CreateTripScheduleInput{
			TripID:        uint(tripID),
			DayNumber:     item.DayNumber,
			SequenceOrder: item.SequenceOrder,
			PlaceName:     item.PlaceName,
			PlaceID:       item.PlaceID,
			Latitude:      item.Latitude,
			Longitude:     item.Longitude,
			StartTime:     item.StartTime,
			EndTime:       item.EndTime,
			Type:          item.Type,
		}
	}

	created, err := h.svc.CreateTripSchedule(c.Context(), inputs)
	if err != nil {
		return c.Status(400).JSON(dto.APIResponse[any]{
			Status:  400,
			Message: "failed to create trip schedule",
			Error:   err.Error(),
		})
	}

	result := make([]dto.TripScheduleItemDTO, len(created))
	for i, item := range created {
		result[i] = toScheduleItemDTO(item)
	}

	return c.Status(201).JSON(dto.APIResponse[[]dto.TripScheduleItemDTO]{
		Status:  201,
		Message: "trip schedule created successfully",
		Data:    &result,
	})
}
