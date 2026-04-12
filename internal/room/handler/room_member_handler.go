package handler

import (
	"strconv"

	"github.com/gofiber/fiber/v2"
	authMiddleware "github.com/manatsanan0209/Vibe-Voyage_Backend/internal/auth/middleware"
	"github.com/manatsanan0209/Vibe-Voyage_Backend/internal/domain"
	"github.com/manatsanan0209/Vibe-Voyage_Backend/internal/dto"
)

func (h *roomHandler) GetMembers(c *fiber.Ctx) error {
	roomID, err := strconv.ParseUint(c.Params("roomID"), 10, 64)
	if err != nil {
		return c.Status(400).JSON(dto.APIResponse[any]{
			Status:  400,
			Message: "bad request",
			Error:   "roomID must be a number",
		})
	}

	members, err := h.svc.GetMembersByRoomID(c.Context(), uint(roomID))
	if err != nil {
		return c.Status(500).JSON(dto.APIResponse[any]{
			Status:  500,
			Message: "failed to get room members",
			Error:   err.Error(),
		})
	}

	result := make([]dto.RoomMemberResponseDTO, 0, len(members))
	for _, m := range members {
		result = append(result, dto.RoomMemberResponseDTO{
			RoomMemberID: m.RoomMemberID,
			RoomID:       m.RoomID,
			UserID:       m.UserID,
			Username:     m.User.Username,
			Role:         m.Role,
			RoleName:     roleName(m.Role),
			CreatedAt:    m.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		})
	}

	return c.Status(200).JSON(dto.APIResponse[[]dto.RoomMemberResponseDTO]{
		Status:  200,
		Message: "success",
		Data:    &result,
	})
}

func (h *roomHandler) AddMember(c *fiber.Ctx) error {
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

	members, err := h.svc.GetMembersByRoomID(c.Context(), uint(roomID))
	if err != nil {
		return c.Status(500).JSON(dto.APIResponse[any]{
			Status:  500,
			Message: "failed to validate room permissions",
			Error:   err.Error(),
		})
	}

	isOwner := false
	for _, m := range members {
		if m.UserID == requesterID && m.Role == domain.RoleOwner {
			isOwner = true
			break
		}
	}
	if !isOwner {
		return c.Status(403).JSON(dto.APIResponse[any]{
			Status:  403,
			Message: "forbidden",
			Error:   "only room owner can add members",
		})
	}

	req := new(dto.AddRoomMemberRequestDTO)
	if err := c.BodyParser(req); err != nil {
		return c.Status(400).JSON(dto.APIResponse[any]{
			Status:  400,
			Message: "bad request",
			Error:   "invalid request body",
		})
	}

	if req.UserID == 0 {
		return c.Status(400).JSON(dto.APIResponse[any]{
			Status:  400,
			Message: "bad request",
			Error:   "user_id is required",
		})
	}

	member, err := h.svc.AddMember(c.Context(), uint(roomID), req.UserID)
	if err != nil {
		return c.Status(400).JSON(dto.APIResponse[any]{
			Status:  400,
			Message: "failed to add member",
			Error:   err.Error(),
		})
	}

	result := dto.RoomMemberResponseDTO{
		RoomMemberID: member.RoomMemberID,
		RoomID:       member.RoomID,
		UserID:       member.UserID,
		Username:     member.User.Username,
		Role:         member.Role,
		RoleName:     roleName(member.Role),
		CreatedAt:    member.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}

	return c.Status(201).JSON(dto.APIResponse[dto.RoomMemberResponseDTO]{
		Status:  201,
		Message: "member added successfully",
		Data:    &result,
	})
}

func (h *roomHandler) DeleteMember(c *fiber.Ctx) error {
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

	if err := h.svc.DeleteMember(c.Context(), uint(roomID), requesterID, uint(memberID)); err != nil {
		return c.Status(400).JSON(dto.APIResponse[any]{
			Status:  400,
			Message: "failed to remove member",
			Error:   err.Error(),
		})
	}

	return c.Status(200).JSON(dto.APIResponse[any]{
		Status:  200,
		Message: "member removed successfully",
	})
}
