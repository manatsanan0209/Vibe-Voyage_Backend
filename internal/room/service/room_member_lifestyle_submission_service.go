package service

import (
	"context"
	"encoding/json"

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

	lifestyleByUserID := make(map[uint]domain.UserLifestyle, len(lifestyles))
	for _, lifestyle := range lifestyles {
		lifestyleByUserID[lifestyle.UserID] = lifestyle
	}

	result := make([]domain.MemberLifestyleSubmissionStatus, 0, len(members))
	for _, member := range members {
		lifestyle, hasSubmitted := lifestyleByUserID[member.UserID]

		var submittedLifestyleID *uint
		hasAnalyzed := false
		if hasSubmitted {
			id := lifestyle.LifestyleID
			submittedLifestyleID = &id
			hasAnalyzed = isStructuredLifestyleReady(lifestyle.StructuredLifestyle)
		}

		result = append(result, domain.MemberLifestyleSubmissionStatus{
			RoomMemberID:          member.RoomMemberID,
			RoomID:                member.RoomID,
			UserID:                member.UserID,
			Username:              member.User.Username,
			Role:                  member.Role,
			HasSubmittedLifestyle: hasSubmitted,
			HasAnalyzedLifestyle:  hasAnalyzed,
			SubmittedLifestyleID:  submittedLifestyleID,
		})
	}

	return result, nil
}

func isStructuredLifestyleReady(value *string) bool {
	if value == nil {
		return false
	}

	var payload interface{}
	return json.Unmarshal([]byte(*value), &payload) == nil
}
