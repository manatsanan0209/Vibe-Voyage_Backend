package service

import (
	"context"

	"github.com/manatsanan0209/Vibe-Voyage_Backend/internal/domain"
)

func (s *roomService) ListMemberLifestyleSubmissions(ctx context.Context, roomID, requesterUserID uint) ([]domain.MemberLifestyleSubmissionStatus, error) {
	if _, err := s.getMemberRole(ctx, roomID, requesterUserID); err != nil {
		return nil, err
	}

	members, err := s.memberRepo.GetByRoomID(ctx, roomID)
	if err != nil {
		return nil, err
	}

	lifestyles, err := s.lifestyleRepo.GetByRoomID(ctx, roomID)
	if err != nil {
		return nil, err
	}

	submittedLifestyleByUserID := make(map[uint]uint, len(lifestyles))
	for _, lifestyle := range lifestyles {
		submittedLifestyleByUserID[lifestyle.UserID] = lifestyle.LifestyleID
	}

	result := make([]domain.MemberLifestyleSubmissionStatus, 0, len(members))
	for _, member := range members {
		lifestyleID, hasSubmitted := submittedLifestyleByUserID[member.UserID]

		var submittedLifestyleID *uint
		if hasSubmitted {
			id := lifestyleID
			submittedLifestyleID = &id
		}

		result = append(result, domain.MemberLifestyleSubmissionStatus{
			RoomMemberID:          member.RoomMemberID,
			RoomID:                member.RoomID,
			UserID:                member.UserID,
			Username:              member.User.Username,
			Role:                  member.Role,
			HasSubmittedLifestyle: hasSubmitted,
			SubmittedLifestyleID:  submittedLifestyleID,
		})
	}

	return result, nil
}
