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

	if err := s.lifestyleRepo.DeleteByUserAndRoom(ctx, target.UserID, target.RoomID); err != nil {
		return err
	}

	return s.memberRepo.DeleteMember(ctx, roomMemberID)
}

func (s *roomService) LeaveRoom(ctx context.Context, roomID, userID uint) error {
	members, err := s.memberRepo.GetByRoomID(ctx, roomID)
	if err != nil {
		return err
	}

	var myMemberID uint
	var myRole int
	found := false
	for _, member := range members {
		if member.UserID == userID {
			myMemberID = member.RoomMemberID
			myRole = member.Role
			found = true
			break
		}
	}

	if !found {
		return errors.New("you are not a member of this room")
	}

	if myRole == domain.RoleOwner {
		return errors.New("room owner cannot leave room")
	}

	if err := s.lifestyleRepo.DeleteByUserAndRoom(ctx, userID, roomID); err != nil {
		return err
	}

	return s.memberRepo.DeleteMember(ctx, myMemberID)
}

func (s *roomService) AddMember(ctx context.Context, roomID, userID uint) (*domain.RoomMember, error) {
	return s.addMemberWithRole(ctx, roomID, userID, domain.RoleMember)
}

func (s *roomService) addMemberWithRole(ctx context.Context, roomID, userID uint, role int) (*domain.RoomMember, error) {
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
		Role:   role,
	}
	return s.memberRepo.AddMember(ctx, member)
}
