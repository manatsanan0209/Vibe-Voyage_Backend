package service

import (
	"context"
	"errors"

	"github.com/manatsanan0209/Vibe-Voyage_Backend/internal/domain"
)

type roomMemberService struct {
	repo domain.RoomMemberRepository
}

func NewRoomMemberService(repo domain.RoomMemberRepository) domain.RoomMemberService {
	return &roomMemberService{repo: repo}
}

func (s *roomMemberService) GetMembersByRoomID(ctx context.Context, roomID uint) ([]domain.RoomMember, error) {
	return s.repo.GetByRoomID(ctx, roomID)
}

func (s *roomMemberService) AddMember(ctx context.Context, roomID, userID uint) (*domain.RoomMember, error) {
	exists, err := s.repo.ExistsByRoomAndUser(ctx, roomID, userID)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, errors.New("user is already a member of this room")
	}

	member := &domain.RoomMember{
		RoomID: roomID,
		UserID: userID,
		Role:   domain.RoleMember,
	}
	return s.repo.AddMember(ctx, member)
}
