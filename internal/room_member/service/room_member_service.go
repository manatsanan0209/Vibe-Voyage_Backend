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

func (s *roomMemberService) DeleteMember(ctx context.Context, roomID, requesterUserID, roomMemberID uint) error {
	// ตรวจสอบว่า requester เป็น owner ของ room นี้
	members, err := s.repo.GetByRoomID(ctx, roomID)
	if err != nil {
		return err
	}

	isOwner := false
	for _, m := range members {
		if m.UserID == requesterUserID && m.Role == domain.RoleOwner {
			isOwner = true
			break
		}
	}
	if !isOwner {
		return errors.New("only the room owner can remove members")
	}

	// ตรวจสอบว่า member ที่จะลบอยู่ใน room นี้จริง
	target, err := s.repo.GetByID(ctx, roomMemberID)
	if err != nil {
		return errors.New("member not found")
	}
	if target.RoomID != roomID {
		return errors.New("member does not belong to this room")
	}
	if target.Role == domain.RoleOwner {
		return errors.New("cannot remove the room owner")
	}

	return s.repo.DeleteMember(ctx, roomMemberID)
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
