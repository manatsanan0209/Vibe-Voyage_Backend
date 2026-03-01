package handler

import (
	"encoding/json"
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/manatsanan0209/Vibe-Voyage_Backend/internal/domain"
	"github.com/manatsanan0209/Vibe-Voyage_Backend/internal/dto"
)

type hotelHandler struct {
	svc domain.HotelService
}

func NewHotelHandler(svc domain.HotelService) *hotelHandler {
	return &hotelHandler{svc: svc}
}

func (h *hotelHandler) RegisterRoutes(app *fiber.App) {
	api := app.Group("/api/hotels")
	api.Get("/accommodation-types", h.GetAccommodationTypes)
	api.Get("/price-ranges", h.GetPriceRanges)
	api.Get("/search", h.GetByName)
	api.Get("/", h.List)
	api.Get("/:id", h.GetByID)
}

func (h *hotelHandler) List(c *fiber.Ctx) error {
	limit, _ := strconv.Atoi(c.Query("limit", "20"))
	offset, _ := strconv.Atoi(c.Query("offset", "0"))

	filter := domain.HotelFilter{
		ProvinceID:   c.Query("province_id"),
		DistrictID:   c.Query("district_id"),
		AccomTypeID:  c.Query("accom_type_id"),
		PriceRangeID: c.Query("price_range_id"),
		Search:       c.Query("search"),
		Limit:        limit,
		Offset:       offset,
	}

	hotels, total, err := h.svc.ListHotels(c.Context(), filter)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(dto.APIResponse[any]{
			Status:  fiber.StatusInternalServerError,
			Message: "failed to list hotels",
			Error:   err.Error(),
		})
	}

	type listResult struct {
		Total int64           `json:"total"`
		Items []*domain.Hotel `json:"items"`
	}
	result := listResult{Total: total, Items: hotels}
	return c.Status(fiber.StatusOK).JSON(dto.APIResponse[listResult]{
		Status:  fiber.StatusOK,
		Message: "success",
		Data:    &result,
	})
}

func (h *hotelHandler) GetByID(c *fiber.Ctx) error {
	id := c.Params("id")
	hotel, err := h.svc.GetHotelByID(c.Context(), id)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(dto.APIResponse[any]{
			Status:  fiber.StatusInternalServerError,
			Message: "failed to get hotel",
			Error:   err.Error(),
		})
	}
	if hotel == nil {
		return c.Status(fiber.StatusNotFound).JSON(dto.APIResponse[any]{
			Status:  fiber.StatusNotFound,
			Message: "hotel not found",
			Error:   "not found",
		})
	}
	return c.Status(fiber.StatusOK).JSON(dto.APIResponse[domain.Hotel]{
		Status:  fiber.StatusOK,
		Message: "success",
		Data:    hotel,
	})
}

func (h *hotelHandler) GetAccommodationTypes(c *fiber.Ctx) error {
	types, err := h.svc.GetAccommodationTypes(c.Context())
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(dto.APIResponse[any]{
			Status:  fiber.StatusInternalServerError,
			Message: "failed to get accommodation types",
			Error:   err.Error(),
		})
	}
	return c.Status(fiber.StatusOK).JSON(dto.APIResponse[[]*domain.AccommodationType]{
		Status:  fiber.StatusOK,
		Message: "success",
		Data:    &types,
	})
}

func (h *hotelHandler) GetPriceRanges(c *fiber.Ctx) error {
	ranges, err := h.svc.GetPriceRanges(c.Context())
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(dto.APIResponse[any]{
			Status:  fiber.StatusInternalServerError,
			Message: "failed to get price ranges",
			Error:   err.Error(),
		})
	}
	return c.Status(fiber.StatusOK).JSON(dto.APIResponse[[]*domain.PriceRange]{
		Status:  fiber.StatusOK,
		Message: "success",
		Data:    &ranges,
	})
}

func (h *hotelHandler) GetByName(c *fiber.Ctx) error {
	name := c.Query("name")
	if name == "" {
		return c.Status(fiber.StatusBadRequest).JSON(dto.APIResponse[any]{
			Status:  fiber.StatusBadRequest,
			Message: "name query parameter is required",
			Error:   "missing name",
		})
	}

	hotels, err := h.svc.GetHotelByName(c.Context(), name)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(dto.APIResponse[any]{
			Status:  fiber.StatusInternalServerError,
			Message: "failed to search hotel by name",
			Error:   err.Error(),
		})
	}

	fields := c.Query("fields", "*")
	if fields == "*" {
		return c.Status(fiber.StatusOK).JSON(dto.APIResponse[[]*domain.Hotel]{
			Status:  fiber.StatusOK,
			Message: "success",
			Data:    &hotels,
		})
	}

	fieldSet := make(map[string]bool)
	for _, f := range strings.Split(fields, ",") {
		fieldSet[strings.TrimSpace(f)] = true
	}

	result := make([]map[string]interface{}, 0, len(hotels))
	for _, h := range hotels {
		b, _ := json.Marshal(h)
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
