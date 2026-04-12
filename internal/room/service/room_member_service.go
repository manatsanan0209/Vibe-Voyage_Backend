package service

import (
	"context"
	"errors"

	"github.com/manatsanan0209/Vibe-Voyage_Backend/internal/domain"
)

func (s *roomService) GetMembersByRoomID(ctx context.Context, roomID uint) ([]domain.RoomMember, error) {
	return s.memberRepo.GetByRoomID(ctx, roomID)
}

func (s *roomService) DeleteMember(ctx context.Context, roomID, requesterUserID, roomMemberID uint) error {
	members, err := s.memberRepo.GetByRoomID(ctx, roomID)
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

	target, err := s.memberRepo.GetByID(ctx, roomMemberID)
	if err != nil {
		return errors.New("member not found")
	}
	if target.RoomID != roomID {
		return errors.New("member does not belong to this room")
	}
	if target.Role == domain.RoleOwner {
		return errors.New("cannot remove the room owner")
	}

	return s.memberRepo.DeleteMember(ctx, roomMemberID)
}

func (s *roomService) AddMember(ctx context.Context, roomID, userID uint) (*domain.RoomMember, error) {
	exists, err := s.memberRepo.ExistsByRoomAndUser(ctx, roomID, userID)
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
	return s.memberRepo.AddMember(ctx, member)
}
