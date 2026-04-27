package repository

import (
	"context"

	"github.com/manatsanan0209/Vibe-Voyage_Backend/internal/domain"
	"gorm.io/gorm"
)

type pgNotificationRepository struct {
	db *gorm.DB
}

func NewNotificationRepository(db *gorm.DB) domain.NotificationRepository {
	return &pgNotificationRepository{db: db}
}

func (r *pgNotificationRepository) Create(ctx context.Context, n *domain.Notification) error {
	return r.db.WithContext(ctx).Create(n).Error
}

func (r *pgNotificationRepository) GetByUserID(ctx context.Context, userID uint, unreadOnly bool) ([]domain.Notification, error) {
	var notifications []domain.Notification
	q := r.db.WithContext(ctx).Where("user_id = ?", userID)
	if unreadOnly {
		q = q.Where("is_read = false")
	}
	err := q.Order("created_at DESC").Find(&notifications).Error
	return notifications, err
}

func (r *pgNotificationRepository) MarkAsRead(ctx context.Context, notificationID, userID uint) error {
	return r.db.WithContext(ctx).
		Model(&domain.Notification{}).
		Where("notification_id = ? AND user_id = ?", notificationID, userID).
		Update("is_read", true).Error
}

func (r *pgNotificationRepository) MarkAllAsRead(ctx context.Context, userID uint) error {
	return r.db.WithContext(ctx).
		Model(&domain.Notification{}).
		Where("user_id = ? AND is_read = false", userID).
		Update("is_read", true).Error
}

func (r *pgNotificationRepository) Delete(ctx context.Context, notificationID, userID uint) error {
	return r.db.WithContext(ctx).
		Where("notification_id = ? AND user_id = ?", notificationID, userID).
		Delete(&domain.Notification{}).Error
}
