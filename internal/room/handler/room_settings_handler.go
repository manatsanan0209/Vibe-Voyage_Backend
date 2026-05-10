package handler

import (
	"strconv"

	"github.com/gofiber/fiber/v2"
	authMiddleware "github.com/manatsanan0209/Vibe-Voyage_Backend/internal/auth/middleware"
	"github.com/manatsanan0209/Vibe-Voyage_Backend/internal/domain"
	"github.com/manatsanan0209/Vibe-Voyage_Backend/internal/dto"
)

func (h *roomHandler) UpdateRoom(c *fiber.Ctx) error {
	requesterID, ok := authMiddleware.GetUserID(c)
	if !ok {
		return c.Status(401).JSON(dto.APIResponse[any]{
			Status:  401,
			Message: "unauthorized",
			Error:   "invalid token claims",
		})
	}

	roomID, err := strconv.ParseUint(c.Params("roomID"), 10, 64)
	if err != nil {
		return c.Status(400).JSON(dto.APIResponse[any]{
			Status:  400,
			Message: "bad request",
			Error:   "roomID must be a number",
		})
	}

	req := new(dto.UpdateRoomRequestDTO)
	if err := c.BodyParser(req); err != nil {
		return c.Status(400).JSON(dto.APIResponse[any]{
			Status:  400,
			Message: "bad request",
			Error:   "invalid request body",
		})
	}

	if req.RoomName == nil && req.RoomImage == nil {
		return c.Status(400).JSON(dto.APIResponse[any]{
			Status:  400,
			Message: "bad request",
			Error:   "at least one field (room_name, room_image) must be provided",
		})
	}

	room, err := h.svc.UpdateRoom(c.Context(), uint(roomID), requesterID, domain.UpdateRoomInput{
		RoomName:  req.RoomName,
		RoomImage: req.RoomImage,
	})
	if err != nil {
		return c.Status(400).JSON(dto.APIResponse[any]{
			Status:  400,
			Message: "failed to update room",
			Error:   err.Error(),
		})
	}

	result := dto.UpdateRoomResponseDTO{
		RoomID:    room.RoomID,
		OwnerID:   room.OwnerID,
		RoomName:  room.RoomName,
		RoomImage: room.RoomImage,
		UpdatedAt: room.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}

	return c.Status(200).JSON(dto.APIResponse[dto.UpdateRoomResponseDTO]{
		Status:  200,
		Message: "room updated successfully",
		Data:    &result,
	})
}

func (h *roomHandler) UpdateMemberRole(c *fiber.Ctx) error {
	requesterID, ok := authMiddleware.GetUserID(c)
	if !ok {
		return c.Status(401).JSON(dto.APIResponse[any]{
			Status:  401,
			Message: "unauthorized",
			Error:   "invalid token claims",
		})
	}

	roomID, err := strconv.ParseUint(c.Params("roomID"), 10, 64)
	if err != nil {
		return c.Status(400).JSON(dto.APIResponse[any]{
			Status:  400,
			Message: "bad request",
			Error:   "roomID must be a number",
		})
	}

	memberID, err := strconv.ParseUint(c.Params("memberID"), 10, 64)
	if err != nil {
		return c.Status(400).JSON(dto.APIResponse[any]{
			Status:  400,
			Message: "bad request",
			Error:   "memberID must be a number",
		})
	}

	req := new(dto.UpdateMemberRoleRequestDTO)
	if err := c.BodyParser(req); err != nil {
		return c.Status(400).JSON(dto.APIResponse[any]{
			Status:  400,
			Message: "bad request",
			Error:   "invalid request body",
		})
	}

	member, err := h.svc.UpdateMemberRole(c.Context(), uint(roomID), requesterID, uint(memberID), req.Role)
	if err != nil {
		return c.Status(400).JSON(dto.APIResponse[any]{
			Status:  400,
			Message: "failed to update member role",
			Error:   err.Error(),
		})
	}

	result := dto.RoomMemberResponseDTO{
		RoomMemberID: member.RoomMemberID,
		RoomID:       member.RoomID,
		UserID:       member.UserID,
		Username:     member.User.Username,
		Role:         member.Role,
		RoleName:     domain.RoomRoleName(member.Role),
		CreatedAt:    member.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}

	return c.Status(200).JSON(dto.APIResponse[dto.RoomMemberResponseDTO]{
		Status:  200,
		Message: "member role updated successfully",
		Data:    &result,
	})
}

func (h *roomHandler) TransferOwnership(c *fiber.Ctx) error {
	requesterID, ok := authMiddleware.GetUserID(c)
	if !ok {
		return c.Status(401).JSON(dto.APIResponse[any]{
			Status:  401,
			Message: "unauthorized",
			Error:   "invalid token claims",
		})
	}

	roomID, err := strconv.ParseUint(c.Params("roomID"), 10, 64)
	if err != nil {
		return c.Status(400).JSON(dto.APIResponse[any]{
			Status:  400,
			Message: "bad request",
			Error:   "roomID must be a number",
		})
	}

	req := new(dto.TransferOwnershipRequestDTO)
	if err := c.BodyParser(req); err != nil {
		return c.Status(400).JSON(dto.APIResponse[any]{
			Status:  400,
			Message: "bad request",
			Error:   "invalid request body",
		})
	}

	if req.NewOwnerUserID == 0 {
		return c.Status(400).JSON(dto.APIResponse[any]{
			Status:  400,
			Message: "bad request",
			Error:   "new_owner_user_id is required",
		})
	}

	if err := h.svc.TransferOwnership(c.Context(), uint(roomID), requesterID, req.NewOwnerUserID); err != nil {
		return c.Status(400).JSON(dto.APIResponse[any]{
			Status:  400,
			Message: "failed to transfer ownership",
			Error:   err.Error(),
		})
	}

	return c.Status(200).JSON(dto.APIResponse[any]{
		Status:  200,
		Message: "ownership transferred successfully",
	})
}
