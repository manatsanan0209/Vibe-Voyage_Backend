package repository

import (
	"context"
	"errors"
	"time"

	"github.com/manatsanan0209/Vibe-Voyage_Backend/internal/domain"
	"gorm.io/gorm"
)

type roomMemberRepository struct {
	db *gorm.DB
}

func NewRoomMemberRepository(db *gorm.DB) domain.RoomMemberRepository {
	return &roomMemberRepository{db: db}
}

func NewRoomInviteCodeRepository(db *gorm.DB) domain.RoomInviteCodeRepository {
	return &roomMemberRepository{db: db}
}

func (r *roomMemberRepository) GetByRoomID(ctx context.Context, roomID uint) ([]domain.RoomMember, error) {
	var members []domain.RoomMember
	if err := r.db.WithContext(ctx).Preload("User").Where("room_id = ?", roomID).Find(&members).Error; err != nil {
		return nil, err
	}
	return members, nil
}

func (r *roomMemberRepository) GetRoomsByUserID(ctx context.Context, userID uint) ([]domain.UserRoomSummary, error) {
	type userRoomSummaryRow struct {
		RoomID        uint
		RoomName      string
		RoomImage     string
		OwnerID       uint
		OwnerUsername string
		Role          int
		JoinedAt      time.Time
		MembersCount  int64
	}

	var rows []userRoomSummaryRow
	err := r.db.WithContext(ctx).
		Table("room_members rm").
		Select(`
			r.room_id AS room_id,
			r.room_name AS room_name,
			r.room_image AS room_image,
			r.owner_id AS owner_id,
			owner.username AS owner_username,
			rm.role AS role,
			rm.created_at AS joined_at,
			COUNT(rm_all.room_member_id) AS members_count
		`).
		Joins("JOIN rooms r ON r.room_id = rm.room_id AND r.deleted_at IS NULL").
		Joins("JOIN users owner ON owner.user_id = r.owner_id AND owner.deleted_at IS NULL").
		Joins("LEFT JOIN room_members rm_all ON rm_all.room_id = r.room_id AND rm_all.deleted_at IS NULL").
		Where("rm.user_id = ? AND rm.deleted_at IS NULL", userID).
		Group("r.room_id, r.room_name, r.room_image, r.owner_id, owner.username, rm.role, rm.created_at").
		Order("rm.created_at DESC").
		Scan(&rows).Error
	if err != nil {
		return nil, err
	}

	result := make([]domain.UserRoomSummary, 0, len(rows))
	for _, row := range rows {
		result = append(result, domain.UserRoomSummary{
			RoomID:        row.RoomID,
			RoomName:      row.RoomName,
			RoomImage:     row.RoomImage,
			OwnerID:       row.OwnerID,
			OwnerUsername: row.OwnerUsername,
			Role:          row.Role,
			JoinedAt:      row.JoinedAt,
			MembersCount:  row.MembersCount,
		})
	}

	return result, nil
}

func (r *roomMemberRepository) AddMember(ctx context.Context, member *domain.RoomMember) (*domain.RoomMember, error) {
	if err := r.db.WithContext(ctx).Create(member).Error; err != nil {
		return nil, err
	}
	if err := r.db.WithContext(ctx).Preload("User").First(member, member.RoomMemberID).Error; err != nil {
		return nil, err
	}
	return member, nil
}

func (r *roomMemberRepository) GetByID(ctx context.Context, roomMemberID uint) (*domain.RoomMember, error) {
	var member domain.RoomMember
	if err := r.db.WithContext(ctx).First(&member, roomMemberID).Error; err != nil {
		return nil, err
	}
	return &member, nil
}

func (r *roomMemberRepository) DeleteMember(ctx context.Context, roomMemberID uint) error {
	return r.db.WithContext(ctx).Delete(&domain.RoomMember{}, roomMemberID).Error
}

func (r *roomMemberRepository) ExistsByRoomAndUser(ctx context.Context, roomID, userID uint) (bool, error) {
	var count int64
	if err := r.db.WithContext(ctx).Model(&domain.RoomMember{}).
		Where("room_id = ? AND user_id = ?", roomID, userID).
		Count(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
}

func (r *roomMemberRepository) Create(ctx context.Context, invite *domain.RoomInviteCode) error {
	return r.db.WithContext(ctx).Create(invite).Error
}

func (r *roomMemberRepository) GetByCode(ctx context.Context, code string) (*domain.RoomInviteCode, error) {
	var invite domain.RoomInviteCode
	if err := r.db.WithContext(ctx).Where("invite_code = ?", code).First(&invite).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &invite, nil
}

func (r *roomMemberRepository) ListByRoomID(ctx context.Context, roomID uint) ([]domain.RoomInviteCode, error) {
	var invites []domain.RoomInviteCode
	if err := r.db.WithContext(ctx).
		Where("room_id = ?", roomID).
		Order("created_at DESC").
		Find(&invites).Error; err != nil {
		return nil, err
	}
	return invites, nil
}

func (r *roomMemberRepository) ExistsCode(ctx context.Context, code string) (bool, error) {
	var count int64
	if err := r.db.WithContext(ctx).Model(&domain.RoomInviteCode{}).
		Where("invite_code = ?", code).
		Count(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
}
