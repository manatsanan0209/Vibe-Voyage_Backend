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

type roomMemberHandler struct {
	svc domain.RoomMemberService
}

func NewRoomMemberHandler(svc domain.RoomMemberService) *roomMemberHandler {
	return &roomMemberHandler{svc: svc}
}

func (h *roomMemberHandler) RegisterRoutes(app *fiber.App) {
	api := app.Group("/api/rooms", authMiddleware.Authorize())
	api.Get("/:roomID/members", h.GetMembers)
	api.Post("/:roomID/members", h.AddMember)
	api.Delete("/:roomID/members/:memberID", h.DeleteMember)
	api.Get("/:roomID/invite-codes", h.ListInviteCodes)
	api.Post("/:roomID/invite-codes", h.CreateInviteCode)
	api.Post("/join-by-invite-code", h.JoinByInviteCode)
}

func roleName(role int) string {
	switch role {
	case domain.RoleOwner:
		return "owner"
	case domain.RoleMember:
		return "member"
	default:
		return "unknown"
	}
}

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

func (h *roomMemberHandler) GetMembers(c *fiber.Ctx) error {
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

func (h *roomMemberHandler) AddMember(c *fiber.Ctx) error {
	roomID, err := strconv.ParseUint(c.Params("roomID"), 10, 64)
	if err != nil {
		return c.Status(400).JSON(dto.APIResponse[any]{
			Status:  400,
			Message: "bad request",
			Error:   "roomID must be a number",
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

func (h *roomMemberHandler) DeleteMember(c *fiber.Ctx) error {
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

func (h *roomMemberHandler) CreateInviteCode(c *fiber.Ctx) error {
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

func (h *roomMemberHandler) ListInviteCodes(c *fiber.Ctx) error {
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

func (h *roomMemberHandler) JoinByInviteCode(c *fiber.Ctx) error {
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
