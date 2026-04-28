package handler

import (
	"errors"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
	authMiddleware "github.com/manatsanan0209/Vibe-Voyage_Backend/internal/auth/middleware"
	"github.com/manatsanan0209/Vibe-Voyage_Backend/internal/domain"
	"github.com/manatsanan0209/Vibe-Voyage_Backend/internal/dto"
)

type tripHandler struct {
	svc           domain.TripService
	suggestionSvc domain.TripSuggestionService
}

func NewTripHandler(svc domain.TripService, suggestionSvc domain.TripSuggestionService) *tripHandler {
	return &tripHandler{svc: svc, suggestionSvc: suggestionSvc}
}

func (h *tripHandler) RegisterRoutes(app *fiber.App) {
	api := app.Group("/api/trip")
	api.Use(authMiddleware.Authorize())
	api.Post("/", h.CreateTrip)
	api.Post("/join-by-invite-code", h.JoinTripByInviteCode)
	api.Get("/:tripID/schedule", h.GetTripSchedule)
	api.Post("/:tripID/schedule", h.CreateTripSchedule)
	api.Put("/:tripID/schedule", h.ReplaceTripSchedule)
	api.Post("/:tripID/reschedule", h.RescheduleTrip)
	api.Get("/:tripID/publish", h.GetPublishStatus)
	api.Post("/:tripID/publish", h.PublishTrip)
	api.Delete("/:tripID/publish", h.UnpublishTrip)
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
	userID, ok := authMiddleware.GetUserID(c)
	if !ok {
		return c.Status(401).JSON(dto.APIResponse[any]{
			Status:  401,
			Message: "unauthorized",
			Error:   "invalid token claims",
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

	result, err := h.svc.GetTripSchedule(c.Context(), userID, uint(tripID))
	if err != nil {
		if errors.Is(err, domain.ErrForbidden) {
			return c.Status(403).JSON(dto.APIResponse[any]{
				Status:  403,
				Message: "forbidden",
				Error:   "you do not have access to this trip",
			})
		}
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
		TripID:          result.Trip.TripID,
		DestinationName: result.Trip.DestinationName,
		StartDate:       result.Trip.StartDate.Format("2006-01-02"),
		EndDate:         result.Trip.EndDate.Format("2006-01-02"),
		Suggestions:     suggestions,
		Days:            days,
	}

	if pt, err := h.suggestionSvc.GetPublishedTripByTripID(c.Context(), result.Trip.TripID); err == nil {
		resp.IsPublished = true
		id := pt.PublishedTripID
		resp.PublishedTripID = &id
	}

	return c.Status(200).JSON(dto.APIResponse[dto.GetTripScheduleResponseDTO]{
		Status:  200,
		Message: "success",
		Data:    &resp,
	})
}

func (h *tripHandler) CreateTrip(c *fiber.Ctx) error {
	userID, ok := authMiddleware.GetUserID(c)
	if !ok {
		return c.Status(401).JSON(dto.APIResponse[any]{
			Status:  401,
			Message: "unauthorized",
			Error:   "invalid token claims",
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
			Latitude:        d.Latitude,
			Longitude:       d.Longitude,
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

	result, err := h.svc.CreateTrip(c.Context(), userID, input)
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

	suggestions := make([]dto.TripScheduleItemDTO, 0, len(result.Suggestions))
	for _, item := range result.Suggestions {
		suggestions = append(suggestions, toScheduleItemDTO(item))
	}
	resp.Suggestions = suggestions

	return c.Status(201).JSON(dto.APIResponse[dto.CreateTripResponseDTO]{
		Status:  201,
		Message: "trip created successfully",
		Data:    &resp,
	})
}

func (h *tripHandler) JoinTripByInviteCode(c *fiber.Ctx) error {
	userID, ok := authMiddleware.GetUserID(c)
	if !ok {
		return c.Status(401).JSON(dto.APIResponse[any]{
			Status:  401,
			Message: "unauthorized",
			Error:   "invalid token claims",
		})
	}

	req := new(dto.JoinTripByInviteCodeRequestDTO)
	if err := c.BodyParser(req); err != nil {
		return c.Status(400).JSON(dto.APIResponse[any]{
			Status:  400,
			Message: "bad request",
			Error:   "invalid request body",
		})
	}

	result, err := h.svc.JoinTripByInviteCode(c.Context(), userID, req.InviteCode)
	if err != nil {
		return c.Status(400).JSON(dto.APIResponse[any]{
			Status:  400,
			Message: "failed to join trip",
			Error:   err.Error(),
		})
	}

	resp := dto.JoinTripByInviteCodeResponseDTO{
		TripID:          result.Trip.TripID,
		RoomID:          result.Trip.RoomID,
		DestinationName: result.Trip.DestinationName,
		StartDate:       result.Trip.StartDate.Format("2006-01-02"),
		EndDate:         result.Trip.EndDate.Format("2006-01-02"),
		RoomMemberID:    result.Member.RoomMemberID,
		UserID:          result.Member.UserID,
		Username:        result.Member.User.Username,
		Role:            result.Member.Role,
		RoleName:        domain.RoomRoleName(result.Member.Role),
		JoinedAt:        result.Member.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}

	return c.Status(201).JSON(dto.APIResponse[dto.JoinTripByInviteCodeResponseDTO]{
		Status:  201,
		Message: "joined trip successfully",
		Data:    &resp,
	})
}

func (h *tripHandler) CreateTripSchedule(c *fiber.Ctx) error {
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

func (h *tripHandler) ReplaceTripSchedule(c *fiber.Ctx) error {
	userID, ok := authMiddleware.GetUserID(c)
	if !ok {
		return c.Status(401).JSON(dto.APIResponse[any]{
			Status:  401,
			Message: "unauthorized",
			Error:   "invalid token claims",
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
			TripScheduleID: item.TripScheduleID,
			TripID:         uint(tripID),
			DayNumber:      item.DayNumber,
			SequenceOrder:  item.SequenceOrder,
			PlaceName:      item.PlaceName,
			PlaceID:        item.PlaceID,
			Latitude:       item.Latitude,
			Longitude:      item.Longitude,
			StartTime:      item.StartTime,
			EndTime:        item.EndTime,
			Type:           item.Type,
		}
	}

	replaced, err := h.svc.ReplaceTripSchedule(c.Context(), userID, uint(tripID), inputs)
	if err != nil {
		if errors.Is(err, domain.ErrForbidden) {
			return c.Status(403).JSON(dto.APIResponse[any]{
				Status:  403,
				Message: "forbidden",
				Error:   "you do not have permission to edit this trip schedule",
			})
		}

		return c.Status(400).JSON(dto.APIResponse[any]{
			Status:  400,
			Message: "failed to replace trip schedule",
			Error:   err.Error(),
		})
	}

	result := make([]dto.TripScheduleItemDTO, len(replaced))
	for i, item := range replaced {
		result[i] = toScheduleItemDTO(item)
	}

	return c.Status(200).JSON(dto.APIResponse[[]dto.TripScheduleItemDTO]{
		Status:  200,
		Message: "trip schedule replaced successfully",
		Data:    &result,
	})
}

func (h *tripHandler) RescheduleTrip(c *fiber.Ctx) error {
	userID, ok := authMiddleware.GetUserID(c)
	if !ok {
		return c.Status(401).JSON(dto.APIResponse[any]{
			Status:  401,
			Message: "unauthorized",
			Error:   "invalid token claims",
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

	result, err := h.svc.RescheduleTrip(c.Context(), userID, uint(tripID))
	if err != nil {
		var notReadyErr *domain.RescheduleAnalysisNotReadyError
		if errors.As(err, &notReadyErr) {
			notReady := make([]dto.RescheduleNotReadyMemberDTO, 0, len(notReadyErr.NotReadyMembers))
			for _, item := range notReadyErr.NotReadyMembers {
				notReady = append(notReady, dto.RescheduleNotReadyMemberDTO{
					UserID:      item.UserID,
					Username:    item.Username,
					LifestyleID: item.LifestyleID,
				})
			}

			conflict := dto.RescheduleConflictResponseDTO{
				NotReadyMembers: notReady,
			}
			return c.Status(409).JSON(dto.APIResponse[dto.RescheduleConflictResponseDTO]{
				Status:  409,
				Message: "reschedule blocked: lifestyle analysis is incomplete",
				Data:    &conflict,
				Error:   "analysis_incomplete",
			})
		}

		if errors.Is(err, domain.ErrForbidden) {
			return c.Status(403).JSON(dto.APIResponse[any]{
				Status:  403,
				Message: "forbidden",
				Error:   "only room owner can reschedule this trip",
			})
		}
		if errors.Is(err, domain.ErrRescheduleConcurrentModification) {
			return c.Status(409).JSON(dto.APIResponse[any]{
				Status:  409,
				Message: "reschedule conflict: another reschedule is currently in progress",
				Error:   err.Error(),
			})
		}

		return c.Status(400).JSON(dto.APIResponse[any]{
			Status:  400,
			Message: "failed to reschedule trip",
			Error:   err.Error(),
		})
	}

	scoreboard := make([]dto.RescheduleTripMemberScoreDTO, 0, len(result.Scoreboard))
	for _, item := range result.Scoreboard {
		scoreboard = append(scoreboard, dto.RescheduleTripMemberScoreDTO{
			UserID:         item.UserID,
			Username:       item.Username,
			Score:          item.Score,
			EffectiveScore: item.EffectiveScore,
			TimesServed:    item.TimesServed,
			DeferredCount:  item.DeferredCount,
		})
	}

	resp := dto.RescheduleTripResponseDTO{
		TripID:           result.TripID,
		ScheduledCount:   result.ScheduledCount,
		SuggestionsCount: result.SuggestionsCount,
		RoundCount:       result.RoundCount,
		SelectedPlaceIDs: result.SelectedPlaceIDs,
		Scoreboard:       scoreboard,
	}

	return c.Status(200).JSON(dto.APIResponse[dto.RescheduleTripResponseDTO]{
		Status:  200,
		Message: "trip rescheduled successfully",
		Data:    &resp,
	})
}

func (h *tripHandler) GetPublishStatus(c *fiber.Ctx) error {
	userID, ok := authMiddleware.GetUserID(c)
	if !ok {
		return c.Status(401).JSON(dto.APIResponse[any]{
			Status:  401,
			Message: "unauthorized",
			Error:   "invalid token claims",
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

	inRoom, err := h.svc.GetTripSchedule(c.Context(), userID, uint(tripID))
	if err != nil {
		if err.Error() == "forbidden" {
			return c.Status(403).JSON(dto.APIResponse[any]{
				Status:  403,
				Message: "forbidden",
				Error:   "you do not have access to this trip",
			})
		}
		return c.Status(500).JSON(dto.APIResponse[any]{
			Status:  500,
			Message: "failed to get trip",
			Error:   err.Error(),
		})
	}
	_ = inRoom

	pt, err := h.suggestionSvc.GetPublishedTripByTripID(c.Context(), uint(tripID))
	if err != nil {
		return c.Status(200).JSON(dto.APIResponse[dto.PublishStatusResponseDTO]{
			Status:  200,
			Message: "success",
			Data: &dto.PublishStatusResponseDTO{
				IsPublished: false,
			},
		})
	}

	resp := dto.PublishStatusResponseDTO{
		IsPublished:     true,
		PublishedTripID: &pt.PublishedTripID,
		Title:           pt.Title,
		Description:     pt.Description,
		ViewCount:       pt.ViewCount,
		LikeCount:       pt.LikeCount,
		PublishedAt:     pt.CreatedAt.Format(time.RFC3339),
	}

	return c.Status(200).JSON(dto.APIResponse[dto.PublishStatusResponseDTO]{
		Status:  200,
		Message: "success",
		Data:    &resp,
	})
}

func (h *tripHandler) PublishTrip(c *fiber.Ctx) error {
	userID, ok := authMiddleware.GetUserID(c)
	if !ok {
		return c.Status(401).JSON(dto.APIResponse[any]{
			Status:  401,
			Message: "unauthorized",
			Error:   "invalid token claims",
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

	req := new(dto.PublishTripRequestDTO)
	if err := c.BodyParser(req); err != nil {
		return c.Status(400).JSON(dto.APIResponse[any]{
			Status:  400,
			Message: "bad request",
			Error:   "invalid request body",
		})
	}

	pt, err := h.suggestionSvc.PublishTrip(c.Context(), uint(tripID), userID, req.Title, req.Description)
	if err != nil {
		if err.Error() == "forbidden" {
			return c.Status(403).JSON(dto.APIResponse[any]{
				Status:  403,
				Message: "forbidden",
				Error:   "only the trip owner can publish",
			})
		}
		if err.Error() == "trip already published" {
			return c.Status(409).JSON(dto.APIResponse[any]{
				Status:  409,
				Message: "conflict",
				Error:   "trip already published",
			})
		}
		return c.Status(400).JSON(dto.APIResponse[any]{
			Status:  400,
			Message: "failed to publish trip",
			Error:   err.Error(),
		})
	}

	resp := dto.PublishTripResponseDTO{
		PublishedTripID: pt.PublishedTripID,
		TripID:          pt.TripID,
		Title:           pt.Title,
		Description:     pt.Description,
		PublishedAt:     pt.CreatedAt.Format(time.RFC3339),
	}

	return c.Status(201).JSON(dto.APIResponse[dto.PublishTripResponseDTO]{
		Status:  201,
		Message: "trip published successfully",
		Data:    &resp,
	})
}

func (h *tripHandler) UnpublishTrip(c *fiber.Ctx) error {
	userID, ok := authMiddleware.GetUserID(c)
	if !ok {
		return c.Status(401).JSON(dto.APIResponse[any]{
			Status:  401,
			Message: "unauthorized",
			Error:   "invalid token claims",
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

	if err := h.suggestionSvc.UnpublishTrip(c.Context(), uint(tripID), userID); err != nil {
		if err.Error() == "forbidden" {
			return c.Status(403).JSON(dto.APIResponse[any]{
				Status:  403,
				Message: "forbidden",
				Error:   "only the publisher can unpublish",
			})
		}
		if err.Error() == "trip is not published" {
			return c.Status(404).JSON(dto.APIResponse[any]{
				Status:  404,
				Message: "not found",
				Error:   "trip is not published",
			})
		}
		return c.Status(500).JSON(dto.APIResponse[any]{
			Status:  500,
			Message: "failed to unpublish trip",
			Error:   err.Error(),
		})
	}

	return c.Status(200).JSON(dto.APIResponse[any]{
		Status:  200,
		Message: "trip unpublished successfully",
	})
}