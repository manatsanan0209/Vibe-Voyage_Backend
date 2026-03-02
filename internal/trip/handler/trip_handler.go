package handler

import (
	"os"
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
		TravelVibes:           req.TravelVibes,
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
