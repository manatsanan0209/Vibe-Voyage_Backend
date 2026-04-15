package service

import (
	"context"
	"time"

	"github.com/manatsanan0209/Vibe-Voyage_Backend/internal/domain"
)

type roomService struct {
	memberRepo       domain.RoomRepository
	inviteRepo       domain.RoomInviteCodeRepository
	lifestyleRepo    domain.UserLifestyleRepository
	userLifestyleSvc domain.UserLifestyleService
	analyzeSemaphore chan struct{}
	analyzeTimeout   time.Duration
}

func NewRoomService(memberRepo domain.RoomRepository, inviteRepo domain.RoomInviteCodeRepository, lifestyleRepo domain.UserLifestyleRepository, userLifestyleSvc domain.UserLifestyleService) domain.RoomService {
	return &roomService{
		memberRepo:       memberRepo,
		inviteRepo:       inviteRepo,
		lifestyleRepo:    lifestyleRepo,
		userLifestyleSvc: userLifestyleSvc,
		analyzeSemaphore: make(chan struct{}, 5),
		analyzeTimeout:   45 * time.Second,
	}
}

func (s *roomService) GetRoomsByUserID(ctx context.Context, userID uint) ([]domain.UserRoomSummary, error) {
	return s.memberRepo.GetRoomsByUserID(ctx, userID)
}
