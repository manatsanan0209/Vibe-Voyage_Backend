package service

import (
	"context"
	"errors"
	"log"

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

	if err := s.memberRepo.DeleteMember(ctx, roomMemberID); err != nil {
		return err
	}

	go s.notifyMembers(members, target.UserID, roomID, domain.NotifTypeMemberLeft, "Member left", "A member has left the room.")
	return nil
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

	if err := s.memberRepo.DeleteMember(ctx, myMemberID); err != nil {
		return err
	}

	go s.notifyMembers(members, userID, roomID, domain.NotifTypeMemberLeft, "Member left", "A member has left the room.")
	return nil
}

func (s *roomService) AddMember(ctx context.Context, roomID, userID uint) (*domain.RoomMember, error) {
	added, err := s.addMemberWithRole(ctx, roomID, userID, domain.RoleMember)
	if err != nil {
		return nil, err
	}

	// notify the added user directly
	go func() {
		refType := "room"
		if err := s.notifSvc.Notify(context.Background(), userID, domain.NotifTypeRoomInvite, "You were added to a room", "You have been added to a room.", &roomID, &refType); err != nil {
			log.Printf("[Notification] room_invite failed (user_id=%d): %v", userID, err)
		}
	}()

	return added, nil
}

func (s *roomService) addMemberWithRole(ctx context.Context, roomID, userID uint, role int) (*domain.RoomMember, error) {
	existingMembers, err := s.memberRepo.GetByRoomID(ctx, roomID)
	if err != nil {
		return nil, err
	}

	for _, m := range existingMembers {
		if m.UserID == userID {
			return nil, errors.New("user is already a member of this room")
		}
	}

	member := &domain.RoomMember{
		RoomID: roomID,
		UserID: userID,
		Role:   role,
	}
	added, err := s.memberRepo.AddMember(ctx, member)
	if err != nil {
		return nil, err
	}

	go s.notifyMembers(existingMembers, userID, roomID, domain.NotifTypeMemberJoined, "New member joined", "A new member has joined the room.")
	return added, nil
}

// notifyMembers sends a notification to all members except excludeUserID.
func (s *roomService) notifyMembers(members []domain.RoomMember, excludeUserID, roomID uint, notifType, title, message string) {
	refType := "room"
	for _, m := range members {
		if m.UserID == excludeUserID {
			continue
		}
		if err := s.notifSvc.Notify(context.Background(), m.UserID, notifType, title, message, &roomID, &refType); err != nil {
			log.Printf("[Notification] %s failed (user_id=%d): %v", notifType, m.UserID, err)
		}
	}
}
