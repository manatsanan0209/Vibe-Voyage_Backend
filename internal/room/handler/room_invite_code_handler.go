package handler

import (
	"strconv"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	authMiddleware "github.com/manatsanan0209/Vibe-Voyage_Backend/internal/auth/middleware"
	"github.com/manatsanan0209/Vibe-Voyage_Backend/internal/domain"
	"github.com/manatsanan0209/Vibe-Voyage_Backend/internal/dto"
)

func inviteToDTO(invite domain.RoomInviteCode) dto.RoomInviteCodeResponseDTO {
	return dto.RoomInviteCodeResponseDTO{
		RoomInviteID:        invite.RoomInviteID,
		RoomID:              invite.RoomID,
		InviteCodeCreatorID: invite.InviteCodeCreatorID,
		InviteCode:          invite.InviteCode,
		Access:              invite.Access,
		ExpireTime:          invite.ExpireTime.Format(time.RFC3339),
		CreatedAt:           invite.CreatedAt.Format(time.RFC3339),
	}
}

func parseExpireTime(value string) (time.Time, error) {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return time.Time{}, nil
	}

	if t, err := time.Parse(time.RFC3339, trimmed); err == nil {
		return t, nil
	}
	if t, err := time.Parse("2006-01-02 15:04:05", trimmed); err == nil {
		return t, nil
	}

	return time.Time{}, fiber.NewError(fiber.StatusBadRequest, "expire_time must be RFC3339 or YYYY-MM-DD HH:MM:SS")
}

func (h *roomHandler) CreateInviteCode(c *fiber.Ctx) error {
	creatorID, ok := authMiddleware.GetUserID(c)
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

	req := new(dto.CreateRoomInviteCodeRequestDTO)
	if err := c.BodyParser(req); err != nil {
		return c.Status(400).JSON(dto.APIResponse[any]{
			Status:  400,
			Message: "bad request",
			Error:   "invalid request body",
		})
	}

	expireTime, err := parseExpireTime(req.ExpireTime)
	if err != nil {
		return c.Status(400).JSON(dto.APIResponse[any]{
			Status:  400,
			Message: "bad request",
			Error:   err.Error(),
		})
	}

	invite, err := h.svc.CreateInviteCode(c.Context(), uint(roomID), creatorID, req.Access, expireTime)
	if err != nil {
		return c.Status(400).JSON(dto.APIResponse[any]{
			Status:  400,
			Message: "failed to create invite code",
			Error:   err.Error(),
		})
	}

	resp := inviteToDTO(*invite)
	return c.Status(201).JSON(dto.APIResponse[dto.RoomInviteCodeResponseDTO]{
		Status:  201,
		Message: "invite code created successfully",
		Data:    &resp,
	})
}

func (h *roomHandler) ListInviteCodes(c *fiber.Ctx) error {
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

	invites, err := h.svc.ListInviteCodes(c.Context(), uint(roomID), requesterID)
	if err != nil {
		return c.Status(400).JSON(dto.APIResponse[any]{
			Status:  400,
			Message: "failed to list invite codes",
			Error:   err.Error(),
		})
	}

	result := make([]dto.RoomInviteCodeResponseDTO, 0, len(invites))
	for _, invite := range invites {
		result = append(result, inviteToDTO(invite))
	}

	return c.Status(200).JSON(dto.APIResponse[[]dto.RoomInviteCodeResponseDTO]{
		Status:  200,
		Message: "success",
		Data:    &result,
	})
}

func (h *roomHandler) JoinByInviteCode(c *fiber.Ctx) error {
	userID, ok := authMiddleware.GetUserID(c)
	if !ok {
		return c.Status(401).JSON(dto.APIResponse[any]{
			Status:  401,
			Message: "unauthorized",
			Error:   "invalid token claims",
		})
	}

	req := new(dto.JoinRoomByInviteCodeRequestDTO)
	if err := c.BodyParser(req); err != nil {
		return c.Status(400).JSON(dto.APIResponse[any]{
			Status:  400,
			Message: "bad request",
			Error:   "invalid request body",
		})
	}

	member, err := h.svc.JoinByInviteCode(c.Context(), userID, req.InviteCode)
	if err != nil {
		return c.Status(400).JSON(dto.APIResponse[any]{
			Status:  400,
			Message: "failed to join room",
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
		Message: "joined room successfully",
		Data:    &result,
	})
}
