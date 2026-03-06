package handler

import (
	"os"
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/manatsanan0209/Vibe-Voyage_Backend/internal/auth/token"
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
	api := app.Group("/api/rooms")
	api.Get("/:roomID/members", h.GetMembers)
	api.Post("/:roomID/members", h.AddMember)
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

	result := make([]dto.RoomMemberDTO, 0, len(members))
	for _, m := range members {
		result = append(result, dto.RoomMemberDTO{
			RoomMemberID: m.RoomMemberID,
			RoomID:       m.RoomID,
			UserID:       m.UserID,
			Role:         m.Role,
			RoleName:     roleName(m.Role),
			CreatedAt:    m.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		})
	}

	return c.Status(200).JSON(dto.APIResponse[[]dto.RoomMemberDTO]{
		Status:  200,
		Message: "success",
		Data:    &result,
	})
}

func (h *roomMemberHandler) AddMember(c *fiber.Ctx) error {
	authHeader := c.Get("Authorization")
	if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
		return c.Status(401).JSON(dto.APIResponse[any]{
			Status:  401,
			Message: "unauthorized",
			Error:   "missing or invalid authorization header",
		})
	}

	tokenStr := strings.TrimPrefix(authHeader, "Bearer ")
	secret := os.Getenv("AUTH_TOKEN_SECRET")
	if _, err := token.Validate(tokenStr, secret); err != nil {
		return c.Status(401).JSON(dto.APIResponse[any]{
			Status:  401,
			Message: "unauthorized",
			Error:   err.Error(),
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

	result := dto.RoomMemberDTO{
		RoomMemberID: member.RoomMemberID,
		RoomID:       member.RoomID,
		UserID:       member.UserID,
		Role:         member.Role,
		RoleName:     roleName(member.Role),
		CreatedAt:    member.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}

	return c.Status(201).JSON(dto.APIResponse[dto.RoomMemberDTO]{
		Status:  201,
		Message: "member added successfully",
		Data:    &result,
	})
}
