package handler

import (
	"strconv"

	"github.com/gofiber/fiber/v2"
	authMiddleware "github.com/manatsanan0209/Vibe-Voyage_Backend/internal/auth/middleware"
	"github.com/manatsanan0209/Vibe-Voyage_Backend/internal/domain"
	"github.com/manatsanan0209/Vibe-Voyage_Backend/internal/dto"
)

type roomHandler struct {
	svc domain.RoomService
}

func NewRoomHandler(svc domain.RoomService) *roomHandler {
	return &roomHandler{svc: svc}
}

func (h *roomHandler) RegisterRoutes(app *fiber.App) {
	api := app.Group("/api/rooms", authMiddleware.Authorize())
	api.Get("/user/:userID", h.GetRoomsByUserID)
	api.Get("/:roomID/members", h.GetMembers)
	api.Get("/:roomID/members/lifestyle-submissions", h.ListMemberLifestyleSubmissions)
	api.Post("/:roomID/members", h.AddMember)
	api.Delete("/:roomID/members/:memberID", h.DeleteMember)
	api.Post("/:roomID/lifestyle", h.AddLifestyle)
	api.Get("/:roomID/invite-codes/history", h.ListInviteCodeHistory)
	api.Get("/:roomID/invite-codes", h.ListInviteCodes)
	api.Post("/:roomID/invite-codes", h.CreateInviteCode)
	api.Post("/join-by-invite-code", h.JoinByInviteCode)
}

func (h *roomHandler) GetRoomsByUserID(c *fiber.Ctx) error {
	requesterID, ok := authMiddleware.GetUserID(c)
	if !ok {
		return c.Status(401).JSON(dto.APIResponse[any]{
			Status:  401,
			Message: "unauthorized",
			Error:   "invalid token claims",
		})
	}

	userID, err := strconv.ParseUint(c.Params("userID"), 10, 64)
	if err != nil {
		return c.Status(400).JSON(dto.APIResponse[any]{
			Status:  400,
			Message: "bad request",
			Error:   "userID must be a number",
		})
	}

	if requesterID != uint(userID) {
		return c.Status(403).JSON(dto.APIResponse[any]{
			Status:  403,
			Message: "forbidden",
			Error:   "cannot access other user's rooms",
		})
	}

	rooms, err := h.svc.GetRoomsByUserID(c.Context(), uint(userID))
	if err != nil {
		return c.Status(500).JSON(dto.APIResponse[any]{
			Status:  500,
			Message: "failed to get user rooms",
			Error:   err.Error(),
		})
	}

	result := make([]dto.UserRoomSummaryResponseDTO, 0, len(rooms))
	for _, room := range rooms {
		result = append(result, dto.UserRoomSummaryResponseDTO{
			RoomID:        room.RoomID,
			TripID:        room.TripID,
			RoomName:      room.RoomName,
			RoomImage:     room.RoomImage,
			OwnerID:       room.OwnerID,
			OwnerUsername: room.OwnerUsername,
			Role:          room.Role,
			RoleName:      domain.RoomRoleName(room.Role),
			JoinedAt:      room.JoinedAt.Format("2006-01-02T15:04:05Z07:00"),
			MembersCount:  room.MembersCount,
		})
	}

	return c.Status(200).JSON(dto.APIResponse[[]dto.UserRoomSummaryResponseDTO]{
		Status:  200,
		Message: "success",
		Data:    &result,
	})
}
