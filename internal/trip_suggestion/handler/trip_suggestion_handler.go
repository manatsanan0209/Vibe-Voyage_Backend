package handler

import (
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
	authMiddleware "github.com/manatsanan0209/Vibe-Voyage_Backend/internal/auth/middleware"
	"github.com/manatsanan0209/Vibe-Voyage_Backend/internal/domain"
	"github.com/manatsanan0209/Vibe-Voyage_Backend/internal/dto"
)

type tripSuggestionHandler struct {
	svc domain.TripSuggestionService
}

func NewTripSuggestionHandler(svc domain.TripSuggestionService) *tripSuggestionHandler {
	return &tripSuggestionHandler{svc: svc}
}

func (h *tripSuggestionHandler) RegisterRoutes(app *fiber.App) {
	api := app.Group("/api/trip-suggestions")
	api.Use(authMiddleware.Authorize())

	api.Get("/", h.GetFeed)
	api.Get("/bookmarks", h.GetBookmarks)
	api.Get("/:publishedTripID", h.GetDetail)
	api.Post("/:publishedTripID/like", h.ToggleLike)
	api.Post("/:publishedTripID/bookmark", h.ToggleBookmark)
	api.Post("/:publishedTripID/use-as-template", h.UseAsTemplate)
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

func toPublishedTripSummaryDTO(meta domain.PublishedTripWithMeta) dto.PublishedTripSummaryDTO {
	return dto.PublishedTripSummaryDTO{
		PublishedTripID: meta.PublishedTrip.PublishedTripID,
		TripID:          meta.PublishedTrip.TripID,
		Title:           meta.PublishedTrip.Title,
		Description:     meta.PublishedTrip.Description,
		DestinationName: meta.Trip.DestinationName,
		DestinationID:   meta.Trip.DestinationID,
		StartDate:       meta.Trip.StartDate.Format("2006-01-02"),
		EndDate:         meta.Trip.EndDate.Format("2006-01-02"),
		ViewCount:       meta.PublishedTrip.ViewCount,
		LikeCount:       meta.PublishedTrip.LikeCount,
		Publisher: dto.PublishedTripPublisherDTO{
			UserID:       meta.PublishedTrip.UserID,
			Username:     meta.PublisherName,
			ProfileImage: meta.PublisherImage,
		},
		IsLiked:      meta.IsLiked,
		IsBookmarked: meta.IsBookmarked,
		PublishedAt:  meta.PublishedTrip.CreatedAt.Format(time.RFC3339),
	}
}

func (h *tripSuggestionHandler) GetFeed(c *fiber.Ctx) error {
	userID, ok := authMiddleware.GetUserID(c)
	if !ok {
		return c.Status(401).JSON(dto.APIResponse[any]{
			Status:  401,
			Message: "unauthorized",
			Error:   "invalid token claims",
		})
	}

	page, _ := strconv.Atoi(c.Query("page", "1"))
	limit, _ := strconv.Atoi(c.Query("limit", "20"))

	trips, total, err := h.svc.GetFeed(c.Context(), page, limit, userID)
	if err != nil {
		return c.Status(500).JSON(dto.APIResponse[any]{
			Status:  500,
			Message: "failed to get trip feed",
			Error:   err.Error(),
		})
	}

	summaries := make([]dto.PublishedTripSummaryDTO, 0, len(trips))
	for _, meta := range trips {
		summaries = append(summaries, toPublishedTripSummaryDTO(meta))
	}

	resp := dto.GetTripFeedResponseDTO{
		Total: total,
		Page:  page,
		Limit: limit,
		Trips: summaries,
	}

	return c.Status(200).JSON(dto.APIResponse[dto.GetTripFeedResponseDTO]{
		Status:  200,
		Message: "success",
		Data:    &resp,
	})
}

func (h *tripSuggestionHandler) GetDetail(c *fiber.Ctx) error {
	userID, ok := authMiddleware.GetUserID(c)
	if !ok {
		return c.Status(401).JSON(dto.APIResponse[any]{
			Status:  401,
			Message: "unauthorized",
			Error:   "invalid token claims",
		})
	}

	publishedTripID, err := strconv.ParseUint(c.Params("publishedTripID"), 10, 64)
	if err != nil {
		return c.Status(400).JSON(dto.APIResponse[any]{
			Status:  400,
			Message: "bad request",
			Error:   "publishedTripID must be a number",
		})
	}

	meta, err := h.svc.GetDetail(c.Context(), uint(publishedTripID), userID)
	if err != nil {
		return c.Status(404).JSON(dto.APIResponse[any]{
			Status:  404,
			Message: "published trip not found",
			Error:   err.Error(),
		})
	}

	days := make([]dto.DayScheduleDTO, 0, len(meta.ScheduleDays))
	for _, day := range meta.ScheduleDays {
		items := make([]dto.TripScheduleItemDTO, 0, len(day.Items))
		for _, item := range day.Items {
			items = append(items, toScheduleItemDTO(item))
		}
		days = append(days, dto.DayScheduleDTO{
			DayNumber: day.DayNumber,
			Items:     items,
		})
	}

	summary := toPublishedTripSummaryDTO(*meta)
	resp := dto.PublishedTripDetailDTO{
		PublishedTripSummaryDTO: summary,
		ScheduleDays:            days,
	}

	return c.Status(200).JSON(dto.APIResponse[dto.PublishedTripDetailDTO]{
		Status:  200,
		Message: "success",
		Data:    &resp,
	})
}

func (h *tripSuggestionHandler) ToggleLike(c *fiber.Ctx) error {
	userID, ok := authMiddleware.GetUserID(c)
	if !ok {
		return c.Status(401).JSON(dto.APIResponse[any]{
			Status:  401,
			Message: "unauthorized",
			Error:   "invalid token claims",
		})
	}

	publishedTripID, err := strconv.ParseUint(c.Params("publishedTripID"), 10, 64)
	if err != nil {
		return c.Status(400).JSON(dto.APIResponse[any]{
			Status:  400,
			Message: "bad request",
			Error:   "publishedTripID must be a number",
		})
	}

	liked, err := h.svc.ToggleLike(c.Context(), uint(publishedTripID), userID)
	if err != nil {
		return c.Status(500).JSON(dto.APIResponse[any]{
			Status:  500,
			Message: "failed to toggle like",
			Error:   err.Error(),
		})
	}

	message := "unliked"
	if liked {
		message = "liked"
	}

	resp := dto.ToggleLikeResponseDTO{Liked: liked}
	return c.Status(200).JSON(dto.APIResponse[dto.ToggleLikeResponseDTO]{
		Status:  200,
		Message: message,
		Data:    &resp,
	})
}

func (h *tripSuggestionHandler) ToggleBookmark(c *fiber.Ctx) error {
	userID, ok := authMiddleware.GetUserID(c)
	if !ok {
		return c.Status(401).JSON(dto.APIResponse[any]{
			Status:  401,
			Message: "unauthorized",
			Error:   "invalid token claims",
		})
	}

	publishedTripID, err := strconv.ParseUint(c.Params("publishedTripID"), 10, 64)
	if err != nil {
		return c.Status(400).JSON(dto.APIResponse[any]{
			Status:  400,
			Message: "bad request",
			Error:   "publishedTripID must be a number",
		})
	}

	bookmarked, err := h.svc.ToggleBookmark(c.Context(), uint(publishedTripID), userID)
	if err != nil {
		return c.Status(500).JSON(dto.APIResponse[any]{
			Status:  500,
			Message: "failed to toggle bookmark",
			Error:   err.Error(),
		})
	}

	message := "unbookmarked"
	if bookmarked {
		message = "bookmarked"
	}

	resp := dto.ToggleBookmarkResponseDTO{Bookmarked: bookmarked}
	return c.Status(200).JSON(dto.APIResponse[dto.ToggleBookmarkResponseDTO]{
		Status:  200,
		Message: message,
		Data:    &resp,
	})
}

func (h *tripSuggestionHandler) GetBookmarks(c *fiber.Ctx) error {
	userID, ok := authMiddleware.GetUserID(c)
	if !ok {
		return c.Status(401).JSON(dto.APIResponse[any]{
			Status:  401,
			Message: "unauthorized",
			Error:   "invalid token claims",
		})
	}

	trips, err := h.svc.GetBookmarks(c.Context(), userID)
	if err != nil {
		return c.Status(500).JSON(dto.APIResponse[any]{
			Status:  500,
			Message: "failed to get bookmarks",
			Error:   err.Error(),
		})
	}

	summaries := make([]dto.PublishedTripSummaryDTO, 0, len(trips))
	for _, meta := range trips {
		summaries = append(summaries, toPublishedTripSummaryDTO(meta))
	}

	return c.Status(200).JSON(dto.APIResponse[[]dto.PublishedTripSummaryDTO]{
		Status:  200,
		Message: "success",
		Data:    &summaries,
	})
}

func (h *tripSuggestionHandler) UseAsTemplate(c *fiber.Ctx) error {
	userID, ok := authMiddleware.GetUserID(c)
	if !ok {
		return c.Status(401).JSON(dto.APIResponse[any]{
			Status:  401,
			Message: "unauthorized",
			Error:   "invalid token claims",
		})
	}

	publishedTripID, err := strconv.ParseUint(c.Params("publishedTripID"), 10, 64)
	if err != nil {
		return c.Status(400).JSON(dto.APIResponse[any]{
			Status:  400,
			Message: "bad request",
			Error:   "publishedTripID must be a number",
		})
	}

	req := new(dto.UseAsTemplateRequestDTO)
	if err := c.BodyParser(req); err != nil {
		return c.Status(400).JSON(dto.APIResponse[any]{
			Status:  400,
			Message: "bad request",
			Error:   "invalid request body",
		})
	}

	if req.RoomName == "" {
		return c.Status(400).JSON(dto.APIResponse[any]{
			Status:  400,
			Message: "bad request",
			Error:   "room_name is required",
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

	input := domain.UseAsTemplateInput{
		RoomName:  req.RoomName,
		RoomImage: req.RoomImage,
		StartDate: startDate,
		EndDate:   endDate,
	}

	result, err := h.svc.UseAsTemplate(c.Context(), uint(publishedTripID), userID, input)
	if err != nil {
		if err.Error() == "cannot use own trip as template" {
			return c.Status(403).JSON(dto.APIResponse[any]{
				Status:  403,
				Message: "forbidden",
				Error:   "you cannot use your own trip as a template",
			})
		}
		return c.Status(400).JSON(dto.APIResponse[any]{
			Status:  400,
			Message: "failed to create trip from template",
			Error:   err.Error(),
		})
	}

	resp := dto.UseAsTemplateResponseDTO{
		RoomID:          result.Room.RoomID,
		TripID:          result.Trip.TripID,
		RoomName:        result.Room.RoomName,
		DestinationName: result.Trip.DestinationName,
		StartDate:       result.Trip.StartDate.Format("2006-01-02"),
		EndDate:         result.Trip.EndDate.Format("2006-01-02"),
	}

	return c.Status(201).JSON(dto.APIResponse[dto.UseAsTemplateResponseDTO]{
		Status:  201,
		Message: "trip created from template successfully",
		Data:    &resp,
	})
}
