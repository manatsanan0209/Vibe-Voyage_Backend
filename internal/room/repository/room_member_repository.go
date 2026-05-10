package repository

import (
	"context"

	"github.com/manatsanan0209/Vibe-Voyage_Backend/internal/domain"
)

func (r *roomRepository) GetByRoomID(ctx context.Context, roomID uint) ([]domain.RoomMember, error) {
	var members []domain.RoomMember
	if err := r.db.WithContext(ctx).Preload("User").Where("room_id = ?", roomID).Find(&members).Error; err != nil {
		return nil, err
	}
	return members, nil
}

func (r *roomRepository) GetMemberLifestyleStatusesByRoomID(ctx context.Context, roomID uint) ([]domain.MemberLifestyleSubmissionStatus, error) {
	members, err := r.GetByRoomID(ctx, roomID)
	if err != nil {
		return nil, err
	}

	var lifestyles []domain.UserLifestyle
	if err := r.db.WithContext(ctx).
		Where("room_id = ?", roomID).
		Find(&lifestyles).Error; err != nil {
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
			hasAnalyzed = domain.IsStructuredLifestyleValid(lifestyle.StructuredLifestyle)
		}

		result = append(result, domain.MemberLifestyleSubmissionStatus{
			RoomMemberID:          member.RoomMemberID,
			RoomID:                member.RoomID,
			UserID:                member.UserID,
			Username:              member.User.Username,
			ProfileImage:          member.User.ProfileImage,
			Role:                  member.Role,
			HasSubmittedLifestyle: hasSubmitted,
			HasAnalyzedLifestyle:  hasAnalyzed,
			SubmittedLifestyleID:  submittedLifestyleID,
		})
	}

	return result, nil
}

func (r *roomRepository) AddMember(ctx context.Context, member *domain.RoomMember) (*domain.RoomMember, error) {
	if err := r.db.WithContext(ctx).Create(member).Error; err != nil {
		return nil, err
	}
	if err := r.db.WithContext(ctx).Preload("User").First(member, member.RoomMemberID).Error; err != nil {
		return nil, err
	}
	return member, nil
}

func (r *roomRepository) GetByID(ctx context.Context, roomMemberID uint) (*domain.RoomMember, error) {
	var member domain.RoomMember
	if err := r.db.WithContext(ctx).First(&member, roomMemberID).Error; err != nil {
		return nil, err
	}
	return &member, nil
}

func (r *roomRepository) DeleteMember(ctx context.Context, roomMemberID uint) error {
	return r.db.WithContext(ctx).Delete(&domain.RoomMember{}, roomMemberID).Error
}

func (r *roomRepository) ExistsByRoomAndUser(ctx context.Context, roomID, userID uint) (bool, error) {
	var count int64
	if err := r.db.WithContext(ctx).Model(&domain.RoomMember{}).
		Where("room_id = ? AND user_id = ?", roomID, userID).
		Count(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
}
