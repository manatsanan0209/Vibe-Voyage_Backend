package handler

import (
	"github.com/gofiber/fiber/v2"
	"github.com/manatsanan0209/Vibe-Voyage_Backend/internal/domain"
	"github.com/manatsanan0209/Vibe-Voyage_Backend/internal/dto"
)

type subdistrictHandler struct {
	svc domain.PlaceService
}

func NewSubdistrictHandler(svc domain.PlaceService) *subdistrictHandler {
	return &subdistrictHandler{svc: svc}
}

func (h *subdistrictHandler) RegisterRoutes(app *fiber.App) {
	api := app.Group("/api/places/subdistricts")
	api.Get("/", h.GetAll)
	api.Get("/district/:district_id", h.GetByDistrictID)
	api.Get("/:id", h.GetByID)
}

func (h *subdistrictHandler) GetAll(c *fiber.Ctx) error {
	subdistricts, err := h.svc.GetAllSubdistricts(c.Context())
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(dto.APIResponse[any]{
			Status:  fiber.StatusInternalServerError,
			Message: "failed to get subdistricts",
			Error:   err.Error(),
		})
	}
	return c.Status(fiber.StatusOK).JSON(dto.APIResponse[[]*domain.Subdistrict]{
		Status:  fiber.StatusOK,
		Message: "success",
		Data:    &subdistricts,
	})
}

func (h *subdistrictHandler) GetByID(c *fiber.Ctx) error {
	id := c.Params("id")
	subdistrict, err := h.svc.GetSubdistrictByID(c.Context(), id)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(dto.APIResponse[any]{
			Status:  fiber.StatusInternalServerError,
			Message: "failed to get subdistrict",
			Error:   err.Error(),
		})
	}
	if subdistrict == nil {
		return c.Status(fiber.StatusNotFound).JSON(dto.APIResponse[any]{
			Status:  fiber.StatusNotFound,
			Message: "subdistrict not found",
			Error:   "not found",
		})
	}
	return c.Status(fiber.StatusOK).JSON(dto.APIResponse[domain.Subdistrict]{
		Status:  fiber.StatusOK,
		Message: "success",
		Data:    subdistrict,
	})
}

func (h *subdistrictHandler) GetByDistrictID(c *fiber.Ctx) error {
	districtID := c.Params("district_id")
	subdistricts, err := h.svc.GetSubdistrictsByDistrict(c.Context(), districtID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(dto.APIResponse[any]{
			Status:  fiber.StatusInternalServerError,
			Message: "failed to get subdistricts by district",
			Error:   err.Error(),
		})
	}
	return c.Status(fiber.StatusOK).JSON(dto.APIResponse[[]*domain.Subdistrict]{
		Status:  fiber.StatusOK,
		Message: "success",
		Data:    &subdistricts,
	})
}
