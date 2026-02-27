package handler

import (
	"github.com/gofiber/fiber/v2"
	"github.com/manatsanan0209/Vibe-Voyage_Backend/internal/domain"
	"github.com/manatsanan0209/Vibe-Voyage_Backend/internal/dto"
)

type regionHandler struct {
	svc domain.PlaceService
}

func NewRegionHandler(svc domain.PlaceService) *regionHandler {
	return &regionHandler{svc: svc}
}

func (h *regionHandler) RegisterRoutes(app *fiber.App) {
	api := app.Group("/api/places/regions")
	api.Get("/", h.GetAll)
	api.Get("/:id", h.GetByID)
}

func (h *regionHandler) GetAll(c *fiber.Ctx) error {
	regions, err := h.svc.GetAllRegions(c.Context())
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(dto.APIResponse[any]{
			Status:  fiber.StatusInternalServerError,
			Message: "failed to get regions",
			Error:   err.Error(),
		})
	}
	return c.Status(fiber.StatusOK).JSON(dto.APIResponse[[]*domain.Region]{
		Status:  fiber.StatusOK,
		Message: "success",
		Data:    &regions,
	})
}

func (h *regionHandler) GetByID(c *fiber.Ctx) error {
	id := c.Params("id")
	region, err := h.svc.GetRegionByID(c.Context(), id)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(dto.APIResponse[any]{
			Status:  fiber.StatusInternalServerError,
			Message: "failed to get region",
			Error:   err.Error(),
		})
	}
	if region == nil {
		return c.Status(fiber.StatusNotFound).JSON(dto.APIResponse[any]{
			Status:  fiber.StatusNotFound,
			Message: "region not found",
			Error:   "not found",
		})
	}
	return c.Status(fiber.StatusOK).JSON(dto.APIResponse[domain.Region]{
		Status:  fiber.StatusOK,
		Message: "success",
		Data:    region,
	})
}
