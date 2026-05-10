package service

import (
	"context"

	"github.com/manatsanan0209/Vibe-Voyage_Backend/internal/domain"
)

func (s *roomService) ListMemberLifestyleSubmissions(ctx context.Context, roomID, requesterUserID uint) ([]domain.MemberLifestyleSubmissionStatus, error) {
	if _, err := s.getMemberRole(ctx, roomID, requesterUserID); err != nil {
		return nil, err
	}

	return s.memberRepo.GetMemberLifestyleStatusesByRoomID(ctx, roomID)
}
