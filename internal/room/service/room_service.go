package service

import (
	"context"

	"github.com/manatsanan0209/Vibe-Voyage_Backend/internal/domain"
)

type roomService struct {
	memberRepo domain.RoomRepository
	inviteRepo domain.RoomInviteCodeRepository
}

func NewRoomService(memberRepo domain.RoomRepository, inviteRepo domain.RoomInviteCodeRepository) domain.RoomService {
	return &roomService{memberRepo: memberRepo, inviteRepo: inviteRepo}
}

func (s *roomService) GetRoomsByUserID(ctx context.Context, userID uint) ([]domain.UserRoomSummary, error) {
	return s.memberRepo.GetRoomsByUserID(ctx, userID)
}
