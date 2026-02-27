package handler

import (
	"encoding/json"
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/manatsanan0209/Vibe-Voyage_Backend/internal/domain"
	"github.com/manatsanan0209/Vibe-Voyage_Backend/internal/dto"
)

type attractionHandler struct {
	svc domain.AttractionService
}

func NewAttractionHandler(svc domain.AttractionService) *attractionHandler {
	return &attractionHandler{svc: svc}
}

func (h *attractionHandler) RegisterRoutes(app *fiber.App) {
	api := app.Group("/api/attractions")
	api.Get("/categories", h.GetCategories)
	api.Get("/types", h.GetTypes)
	api.Get("/search", h.GetByName)
	api.Get("/", h.List)
	api.Get("/:id", h.GetByID)
}

func (h *attractionHandler) List(c *fiber.Ctx) error {
	limit, _ := strconv.Atoi(c.Query("limit", "20"))
	offset, _ := strconv.Atoi(c.Query("offset", "0"))

	filter := domain.AttractionFilter{
		ProvinceID: c.Query("province_id"),
		DistrictID: c.Query("district_id"),
		CategoryID: c.Query("category_id"),
		TypeID:     c.Query("type_id"),
		Search:     c.Query("search"),
		Limit:      limit,
		Offset:     offset,
	}

	attractions, total, err := h.svc.ListAttractions(c.Context(), filter)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(dto.APIResponse[any]{
			Status:  fiber.StatusInternalServerError,
			Message: "failed to list attractions",
			Error:   err.Error(),
		})
	}

	type listResult struct {
		Total int64                `json:"total"`
		Items []*domain.Attraction `json:"items"`
	}
	result := listResult{Total: total, Items: attractions}
	return c.Status(fiber.StatusOK).JSON(dto.APIResponse[listResult]{
		Status:  fiber.StatusOK,
		Message: "success",
		Data:    &result,
	})
}

func (h *attractionHandler) GetByName(c *fiber.Ctx) error {
	name := c.Query("name")
	if name == "" {
		return c.Status(fiber.StatusBadRequest).JSON(dto.APIResponse[any]{
			Status:  fiber.StatusBadRequest,
			Message: "name query parameter is required",
			Error:   "missing name",
		})
	}

	attractions, err := h.svc.GetAttractionByName(c.Context(), name)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(dto.APIResponse[any]{
			Status:  fiber.StatusInternalServerError,
			Message: "failed to search attraction by name",
			Error:   err.Error(),
		})
	}

	fields := c.Query("fields", "*")
	if fields == "*" {
		return c.Status(fiber.StatusOK).JSON(dto.APIResponse[[]*domain.Attraction]{
			Status:  fiber.StatusOK,
			Message: "success",
			Data:    &attractions,
		})
	}

	fieldSet := make(map[string]bool)
	for _, f := range strings.Split(fields, ",") {
		fieldSet[strings.TrimSpace(f)] = true
	}

	result := make([]map[string]interface{}, 0, len(attractions))
	for _, a := range attractions {
		b, _ := json.Marshal(a)
		var m map[string]interface{}
		_ = json.Unmarshal(b, &m)
		filtered := make(map[string]interface{})
		for k, v := range m {
			if fieldSet[k] {
				filtered[k] = v
			}
		}
		result = append(result, filtered)
	}

	return c.Status(fiber.StatusOK).JSON(dto.APIResponse[[]map[string]interface{}]{
		Status:  fiber.StatusOK,
		Message: "success",
		Data:    &result,
	})
}

func (h *attractionHandler) GetByID(c *fiber.Ctx) error {
	id := c.Params("id")
	attraction, err := h.svc.GetAttractionByID(c.Context(), id)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(dto.APIResponse[any]{
			Status:  fiber.StatusInternalServerError,
			Message: "failed to get attraction",
			Error:   err.Error(),
		})
	}
	if attraction == nil {
		return c.Status(fiber.StatusNotFound).JSON(dto.APIResponse[any]{
			Status:  fiber.StatusNotFound,
			Message: "attraction not found",
			Error:   "not found",
		})
	}
	return c.Status(fiber.StatusOK).JSON(dto.APIResponse[domain.Attraction]{
		Status:  fiber.StatusOK,
		Message: "success",
		Data:    attraction,
	})
}

func (h *attractionHandler) GetCategories(c *fiber.Ctx) error {
	categories, err := h.svc.GetAttractionCategories(c.Context())
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(dto.APIResponse[any]{
			Status:  fiber.StatusInternalServerError,
			Message: "failed to get categories",
			Error:   err.Error(),
		})
	}
	return c.Status(fiber.StatusOK).JSON(dto.APIResponse[[]*domain.AttractionCategory]{
		Status:  fiber.StatusOK,
		Message: "success",
		Data:    &categories,
	})
}

func (h *attractionHandler) GetTypes(c *fiber.Ctx) error {
	types, err := h.svc.GetAttractionTypes(c.Context())
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(dto.APIResponse[any]{
			Status:  fiber.StatusInternalServerError,
			Message: "failed to get types",
			Error:   err.Error(),
		})
	}
	return c.Status(fiber.StatusOK).JSON(dto.APIResponse[[]*domain.AttractionType]{
		Status:  fiber.StatusOK,
		Message: "success",
		Data:    &types,
	})
}
