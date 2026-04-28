package handler

import (
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
	authMiddleware "github.com/manatsanan0209/Vibe-Voyage_Backend/internal/auth/middleware"
	"github.com/manatsanan0209/Vibe-Voyage_Backend/internal/domain"
	"github.com/manatsanan0209/Vibe-Voyage_Backend/internal/dto"
)

type userHandler struct {
	svc            domain.UserService
	suggestionSvc  domain.TripSuggestionService
}

func NewUserHandler(svc domain.UserService, suggestionSvc domain.TripSuggestionService) *userHandler {
	return &userHandler{svc: svc, suggestionSvc: suggestionSvc}
}

func (h *userHandler) RegisterRoutes(app *fiber.App) {
	api := app.Group("/api/profile", authMiddleware.Authorize())
	api.Get("/", h.GetProfile)
	api.Patch("/", h.UpdateProfile)
	api.Get("/posts", h.GetMyPosts)
}

func (h *userHandler) GetProfile(c *fiber.Ctx) error {
	userID, ok := authMiddleware.GetUserID(c)
	if !ok {
		return c.Status(401).JSON(dto.APIResponse[any]{
			Status:  401,
			Message: "unauthorized",
			Error:   "invalid token claims",
		})
	}

	user, err := h.svc.GetUserByID(c.Context(), userID)
	if err != nil {
		return c.Status(404).JSON(dto.APIResponse[any]{
			Status:  404,
			Message: "user not found",
			Error:   err.Error(),
		})
	}

	resp := dto.ProfileResponseDTO{
		UserID:       user.UserID,
		Username:     user.Username,
		FullName:     user.FullName,
		Email:        user.Email,
		ProfileImage: user.ProfileImage,
	}

	return c.Status(200).JSON(dto.APIResponse[dto.ProfileResponseDTO]{
		Status:  200,
		Message: "success",
		Data:    &resp,
	})
}

func (h *userHandler) UpdateProfile(c *fiber.Ctx) error {
	userID, ok := authMiddleware.GetUserID(c)
	if !ok {
		return c.Status(401).JSON(dto.APIResponse[any]{
			Status:  401,
			Message: "unauthorized",
			Error:   "invalid token claims",
		})
	}

	req := new(dto.UpdateProfileRequestDTO)
	if err := c.BodyParser(req); err != nil {
		return c.Status(400).JSON(dto.APIResponse[any]{
			Status:  400,
			Message: "bad request",
			Error:   "invalid request body",
		})
	}

	input := domain.UpdateProfileInput{
		Username:     req.Username,
		FullName:     req.FullName,
		ProfileImage: req.ProfileImage,
	}

	user, err := h.svc.UpdateProfile(c.Context(), userID, input)
	if err != nil {
		if err.Error() == "username already taken" {
			return c.Status(409).JSON(dto.APIResponse[any]{
				Status:  409,
				Message: "conflict",
				Error:   err.Error(),
			})
		}
		return c.Status(400).JSON(dto.APIResponse[any]{
			Status:  400,
			Message: "failed to update profile",
			Error:   err.Error(),
		})
	}

	resp := dto.ProfileResponseDTO{
		UserID:       user.UserID,
		Username:     user.Username,
		FullName:     user.FullName,
		Email:        user.Email,
		ProfileImage: user.ProfileImage,
	}

	return c.Status(200).JSON(dto.APIResponse[dto.ProfileResponseDTO]{
		Status:  200,
		Message: "profile updated",
		Data:    &resp,
	})
}

func (h *userHandler) GetMyPosts(c *fiber.Ctx) error {
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

	trips, total, err := h.suggestionSvc.GetMyPosts(c.Context(), userID, page, limit)
	if err != nil {
		return c.Status(500).JSON(dto.APIResponse[any]{
			Status:  500,
			Message: "failed to get posts",
			Error:   err.Error(),
		})
	}

	posts := make([]dto.PublishedTripSummaryDTO, 0, len(trips))
	for _, meta := range trips {
		posts = append(posts, dto.PublishedTripSummaryDTO{
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
		})
	}

	resp := dto.MyPostsResponseDTO{
		Total: total,
		Page:  page,
		Limit: limit,
		Posts: posts,
	}

	return c.Status(200).JSON(dto.APIResponse[dto.MyPostsResponseDTO]{
		Status:  200,
		Message: "success",
		Data:    &resp,
	})
}
