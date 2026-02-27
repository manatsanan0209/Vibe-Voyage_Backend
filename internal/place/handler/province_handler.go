package handler

import (
	"github.com/gofiber/fiber/v2"
	"github.com/manatsanan0209/Vibe-Voyage_Backend/internal/domain"
	"github.com/manatsanan0209/Vibe-Voyage_Backend/internal/dto"
)

type provinceHandler struct {
	svc domain.PlaceService
}

func NewProvinceHandler(svc domain.PlaceService) *provinceHandler {
	return &provinceHandler{svc: svc}
}

func (h *provinceHandler) RegisterRoutes(app *fiber.App) {
	api := app.Group("/api/places/provinces")
	api.Get("/", h.GetAll)
	api.Get("/region/:region_id", h.GetByRegionID)
	api.Get("/:id", h.GetByID)
}

func (h *provinceHandler) GetAll(c *fiber.Ctx) error {
	provinces, err := h.svc.GetAllProvinces(c.Context())
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(dto.APIResponse[any]{
			Status:  fiber.StatusInternalServerError,
			Message: "failed to get provinces",
			Error:   err.Error(),
		})
	}
	return c.Status(fiber.StatusOK).JSON(dto.APIResponse[[]*domain.Province]{
		Status:  fiber.StatusOK,
		Message: "success",
		Data:    &provinces,
	})
}

func (h *provinceHandler) GetByID(c *fiber.Ctx) error {
	id := c.Params("id")
	province, err := h.svc.GetProvinceByID(c.Context(), id)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(dto.APIResponse[any]{
			Status:  fiber.StatusInternalServerError,
			Message: "failed to get province",
			Error:   err.Error(),
		})
	}
	if province == nil {
		return c.Status(fiber.StatusNotFound).JSON(dto.APIResponse[any]{
			Status:  fiber.StatusNotFound,
			Message: "province not found",
			Error:   "not found",
		})
	}
	return c.Status(fiber.StatusOK).JSON(dto.APIResponse[domain.Province]{
		Status:  fiber.StatusOK,
		Message: "success",
		Data:    province,
	})
}

func (h *provinceHandler) GetByRegionID(c *fiber.Ctx) error {
	regionID := c.Params("region_id")
	provinces, err := h.svc.GetProvincesByRegion(c.Context(), regionID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(dto.APIResponse[any]{
			Status:  fiber.StatusInternalServerError,
			Message: "failed to get provinces by region",
			Error:   err.Error(),
		})
	}
	return c.Status(fiber.StatusOK).JSON(dto.APIResponse[[]*domain.Province]{
		Status:  fiber.StatusOK,
		Message: "success",
		Data:    &provinces,
	})
}
