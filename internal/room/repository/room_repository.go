package repository

import (
	"context"
	"time"

	"github.com/manatsanan0209/Vibe-Voyage_Backend/internal/domain"
	"gorm.io/gorm"
)

type roomRepository struct {
	db *gorm.DB
}

func NewRoomRepository(db *gorm.DB) domain.RoomRepository {
	return &roomRepository{db: db}
}

func NewRoomInviteCodeRepository(db *gorm.DB) domain.RoomInviteCodeRepository {
	return &roomRepository{db: db}
}

func (r *roomRepository) GetRoomInfoByID(ctx context.Context, roomID uint) (*domain.Room, error) {
	var room domain.Room
	if err := r.db.WithContext(ctx).First(&room, roomID).Error; err != nil {
		return nil, err
	}
	return &room, nil
}

func (r *roomRepository) UpdateRoom(ctx context.Context, roomID uint, input domain.UpdateRoomInput) (*domain.Room, error) {
	updates := map[string]interface{}{}
	if input.RoomName != nil {
		updates["room_name"] = *input.RoomName
	}
	if input.RoomImage != nil {
		updates["room_image"] = *input.RoomImage
	}
	if err := r.db.WithContext(ctx).Model(&domain.Room{}).Where("room_id = ?", roomID).Updates(updates).Error; err != nil {
		return nil, err
	}
	return r.GetRoomInfoByID(ctx, roomID)
}

func (r *roomRepository) UpdateMemberRole(ctx context.Context, roomMemberID uint, role int) (*domain.RoomMember, error) {
	if err := r.db.WithContext(ctx).Model(&domain.RoomMember{}).
		Where("room_member_id = ?", roomMemberID).
		Update("role", role).Error; err != nil {
		return nil, err
	}
	var member domain.RoomMember
	if err := r.db.WithContext(ctx).Preload("User").First(&member, roomMemberID).Error; err != nil {
		return nil, err
	}
	return &member, nil
}

func (r *roomRepository) TransferOwnership(ctx context.Context, roomID, currentOwnerMemberID, newOwnerMemberID uint) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var newMember domain.RoomMember
		if err := tx.First(&newMember, newOwnerMemberID).Error; err != nil {
			return err
		}

		if err := tx.Model(&domain.RoomMember{}).
			Where("room_member_id = ?", currentOwnerMemberID).
			Update("role", domain.RoleMember).Error; err != nil {
			return err
		}

		if err := tx.Model(&domain.RoomMember{}).
			Where("room_member_id = ?", newOwnerMemberID).
			Update("role", domain.RoleOwner).Error; err != nil {
			return err
		}

		if err := tx.Model(&domain.Room{}).
			Where("room_id = ?", roomID).
			Update("owner_id", newMember.UserID).Error; err != nil {
			return err
		}

		return nil
	})
}

func (r *roomRepository) GetRoomsByUserID(ctx context.Context, userID uint) ([]domain.UserRoomSummary, error) {
	type userRoomSummaryRow struct {
		RoomID        uint
		TripID        uint
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
			COALESCE(t.trip_id, 0) AS trip_id,
			r.room_name AS room_name,
			r.room_image AS room_image,
			r.owner_id AS owner_id,
			owner.username AS owner_username,
			rm.role AS role,
			rm.created_at AS joined_at,
			COUNT(DISTINCT rm_all.room_member_id) AS members_count
		`).
		Joins("JOIN rooms r ON r.room_id = rm.room_id AND r.deleted_at IS NULL").
		Joins("LEFT JOIN trips t ON t.room_id = r.room_id AND t.deleted_at IS NULL").
		Joins("JOIN users owner ON owner.user_id = r.owner_id AND owner.deleted_at IS NULL").
		Joins("LEFT JOIN room_members rm_all ON rm_all.room_id = r.room_id AND rm_all.deleted_at IS NULL").
		Where("rm.user_id = ? AND rm.deleted_at IS NULL", userID).
		Group("r.room_id, t.trip_id, r.room_name, r.room_image, r.owner_id, owner.username, rm.role, rm.created_at").
		Order("rm.created_at DESC").
		Scan(&rows).Error
	if err != nil {
		return nil, err
	}

	result := make([]domain.UserRoomSummary, 0, len(rows))
	for _, row := range rows {
		result = append(result, domain.UserRoomSummary{
			RoomID:        row.RoomID,
			TripID:        row.TripID,
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
