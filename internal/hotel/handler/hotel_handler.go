package handler

import (
	"strconv"

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
