package service

import (
	"context"
	"errors"

	"github.com/manatsanan0209/Vibe-Voyage_Backend/internal/domain"
)

func (s *roomService) UpdateRoom(ctx context.Context, roomID, requesterUserID uint, input domain.UpdateRoomInput) (*domain.Room, error) {
	members, err := s.memberRepo.GetByRoomID(ctx, roomID)
	if err != nil {
		return nil, err
	}

	isOwner := false
	for _, m := range members {
		if m.UserID == requesterUserID && m.Role == domain.RoleOwner {
			isOwner = true
			break
		}
	}
	if !isOwner {
		return nil, errors.New("only the room owner can update room settings")
	}

	if input.RoomName != nil && *input.RoomName == "" {
		return nil, errors.New("room_name cannot be empty")
	}

	return s.memberRepo.UpdateRoom(ctx, roomID, input)
}

func (s *roomService) UpdateMemberRole(ctx context.Context, roomID, requesterUserID, roomMemberID uint, role int) (*domain.RoomMember, error) {
	if role != domain.RoleMember && role != domain.RoleSpectator {
		return nil, errors.New("role must be 2 (member) or 3 (spectator)")
	}

	members, err := s.memberRepo.GetByRoomID(ctx, roomID)
	if err != nil {
		return nil, err
	}

	isOwner := false
	for _, m := range members {
		if m.UserID == requesterUserID && m.Role == domain.RoleOwner {
			isOwner = true
			break
		}
	}
	if !isOwner {
		return nil, errors.New("only the room owner can change member roles")
	}

	target, err := s.memberRepo.GetByID(ctx, roomMemberID)
	if err != nil {
		return nil, errors.New("member not found")
	}
	if target.RoomID != roomID {
		return nil, errors.New("member does not belong to this room")
	}
	if target.Role == domain.RoleOwner {
		return nil, errors.New("cannot change the owner's role; use transfer-ownership instead")
	}

	return s.memberRepo.UpdateMemberRole(ctx, roomMemberID, role)
}

func (s *roomService) TransferOwnership(ctx context.Context, roomID, currentOwnerUserID, newOwnerUserID uint) error {
	if currentOwnerUserID == newOwnerUserID {
		return errors.New("new owner must be a different user")
	}

	members, err := s.memberRepo.GetByRoomID(ctx, roomID)
	if err != nil {
		return err
	}

	var currentOwnerMemberID uint
	var newOwnerMemberID uint
	foundCurrent := false
	foundNew := false

	for _, m := range members {
		if m.UserID == currentOwnerUserID && m.Role == domain.RoleOwner {
			currentOwnerMemberID = m.RoomMemberID
			foundCurrent = true
		}
		if m.UserID == newOwnerUserID {
			newOwnerMemberID = m.RoomMemberID
			foundNew = true
		}
	}

	if !foundCurrent {
		return errors.New("only the room owner can transfer ownership")
	}
	if !foundNew {
		return errors.New("new owner must be an existing room member")
	}

	return s.memberRepo.TransferOwnership(ctx, roomID, currentOwnerMemberID, newOwnerMemberID)
}
