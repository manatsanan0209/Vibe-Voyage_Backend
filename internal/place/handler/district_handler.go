package handler

import (
	"github.com/gofiber/fiber/v2"
	"github.com/manatsanan0209/Vibe-Voyage_Backend/internal/domain"
	"github.com/manatsanan0209/Vibe-Voyage_Backend/internal/dto"
)

type districtHandler struct {
	svc domain.PlaceService
}

func NewDistrictHandler(svc domain.PlaceService) *districtHandler {
	return &districtHandler{svc: svc}
}

func (h *districtHandler) RegisterRoutes(app *fiber.App) {
	api := app.Group("/api/places/districts")
	api.Get("/", h.GetAll)
	api.Get("/province/:province_id", h.GetByProvinceID)
	api.Get("/:id", h.GetByID)
}

func (h *districtHandler) GetAll(c *fiber.Ctx) error {
	districts, err := h.svc.GetAllDistricts(c.Context())
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(dto.APIResponse[any]{
			Status:  fiber.StatusInternalServerError,
			Message: "failed to get districts",
			Error:   err.Error(),
		})
	}
	return c.Status(fiber.StatusOK).JSON(dto.APIResponse[[]*domain.District]{
		Status:  fiber.StatusOK,
		Message: "success",
		Data:    &districts,
	})
}

func (h *districtHandler) GetByID(c *fiber.Ctx) error {
	id := c.Params("id")
	district, err := h.svc.GetDistrictByID(c.Context(), id)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(dto.APIResponse[any]{
			Status:  fiber.StatusInternalServerError,
			Message: "failed to get district",
			Error:   err.Error(),
		})
	}
	if district == nil {
		return c.Status(fiber.StatusNotFound).JSON(dto.APIResponse[any]{
			Status:  fiber.StatusNotFound,
			Message: "district not found",
			Error:   "not found",
		})
	}
	return c.Status(fiber.StatusOK).JSON(dto.APIResponse[domain.District]{
		Status:  fiber.StatusOK,
		Message: "success",
		Data:    district,
	})
}

func (h *districtHandler) GetByProvinceID(c *fiber.Ctx) error {
	provinceID := c.Params("province_id")
	districts, err := h.svc.GetDistrictsByProvince(c.Context(), provinceID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(dto.APIResponse[any]{
			Status:  fiber.StatusInternalServerError,
			Message: "failed to get districts by province",
			Error:   err.Error(),
		})
	}
	return c.Status(fiber.StatusOK).JSON(dto.APIResponse[[]*domain.District]{
		Status:  fiber.StatusOK,
		Message: "success",
		Data:    &districts,
	})
}
