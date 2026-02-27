package handler

import (
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/manatsanan0209/Vibe-Voyage_Backend/internal/domain"
	"github.com/manatsanan0209/Vibe-Voyage_Backend/internal/dto"
)

type restaurantHandler struct {
	svc domain.RestaurantService
}

func NewRestaurantHandler(svc domain.RestaurantService) *restaurantHandler {
	return &restaurantHandler{svc: svc}
}

func (h *restaurantHandler) RegisterRoutes(app *fiber.App) {
	api := app.Group("/api/restaurants")
	api.Get("/food-types", h.GetFoodTypes)
	api.Get("/", h.List)
	api.Get("/:id", h.GetByID)
}

func (h *restaurantHandler) List(c *fiber.Ctx) error {
	limit, _ := strconv.Atoi(c.Query("limit", "20"))
	offset, _ := strconv.Atoi(c.Query("offset", "0"))

	filter := domain.RestaurantFilter{
		ProvinceID: c.Query("province_id"),
		DistrictID: c.Query("district_id"),
		Search:     c.Query("search"),
		Limit:      limit,
		Offset:     offset,
	}

	restaurants, total, err := h.svc.ListRestaurants(c.Context(), filter)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(dto.APIResponse[any]{
			Status:  fiber.StatusInternalServerError,
			Message: "failed to list restaurants",
			Error:   err.Error(),
		})
	}

	type listResult struct {
		Total int64                `json:"total"`
		Items []*domain.Restaurant `json:"items"`
	}
	result := listResult{Total: total, Items: restaurants}
	return c.Status(fiber.StatusOK).JSON(dto.APIResponse[listResult]{
		Status:  fiber.StatusOK,
		Message: "success",
		Data:    &result,
	})
}

func (h *restaurantHandler) GetByID(c *fiber.Ctx) error {
	id := c.Params("id")
	restaurant, err := h.svc.GetRestaurantByID(c.Context(), id)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(dto.APIResponse[any]{
			Status:  fiber.StatusInternalServerError,
			Message: "failed to get restaurant",
			Error:   err.Error(),
		})
	}
	if restaurant == nil {
		return c.Status(fiber.StatusNotFound).JSON(dto.APIResponse[any]{
			Status:  fiber.StatusNotFound,
			Message: "restaurant not found",
			Error:   "not found",
		})
	}
	return c.Status(fiber.StatusOK).JSON(dto.APIResponse[domain.Restaurant]{
		Status:  fiber.StatusOK,
		Message: "success",
		Data:    restaurant,
	})
}

func (h *restaurantHandler) GetFoodTypes(c *fiber.Ctx) error {
	foodTypes, err := h.svc.GetFoodTypes(c.Context())
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(dto.APIResponse[any]{
			Status:  fiber.StatusInternalServerError,
			Message: "failed to get food types",
			Error:   err.Error(),
		})
	}
	return c.Status(fiber.StatusOK).JSON(dto.APIResponse[[]*domain.FoodType]{
		Status:  fiber.StatusOK,
		Message: "success",
		Data:    &foodTypes,
	})
}
