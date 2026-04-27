package service

import (
	"context"
	"errors"

	"github.com/manatsanan0209/Vibe-Voyage_Backend/internal/domain"
	"gorm.io/gorm"
)

type notificationService struct {
	repo        domain.NotificationRepository
	settingsRepo domain.UserSettingsRepository
}

func NewNotificationService(repo domain.NotificationRepository, settingsRepo domain.UserSettingsRepository) domain.NotificationService {
	return &notificationService{repo: repo, settingsRepo: settingsRepo}
}

func (s *notificationService) GetNotifications(ctx context.Context, userID uint, unreadOnly bool) ([]domain.Notification, error) {
	return s.repo.GetByUserID(ctx, userID, unreadOnly)
}

func (s *notificationService) MarkAsRead(ctx context.Context, notificationID, userID uint) error {
	return s.repo.MarkAsRead(ctx, notificationID, userID)
}

func (s *notificationService) MarkAllAsRead(ctx context.Context, userID uint) error {
	return s.repo.MarkAllAsRead(ctx, userID)
}

func (s *notificationService) Delete(ctx context.Context, notificationID, userID uint) error {
	return s.repo.Delete(ctx, notificationID, userID)
}

func (s *notificationService) Notify(ctx context.Context, userID uint, notifType, title, message string, refID *uint, refType *string) error {
	if !s.isEnabled(ctx, userID, notifType) {
		return nil
	}
	n := &domain.Notification{
		UserID:        userID,
		Type:          notifType,
		Title:         title,
		Message:       message,
		ReferenceID:   refID,
		ReferenceType: refType,
	}
	return s.repo.Create(ctx, n)
}

func (s *notificationService) isEnabled(ctx context.Context, userID uint, notifType string) bool {
	settings, err := s.settingsRepo.GetByUserID(ctx, userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return true
		}
		return true
	}
	switch notifType {
	case domain.NotifTypeRoomInvite:
		return settings.NotifyRoomInvite
	case domain.NotifTypeMemberJoined:
		return settings.NotifyMemberJoined
	case domain.NotifTypeMemberLeft:
		return settings.NotifyMemberLeft
	case domain.NotifTypeTripCreated:
		return settings.NotifyTripCreated
	case domain.NotifTypeLifestyleAnalyzed:
		return settings.NotifyLifestyleAnalyzed
	case domain.NotifTypeScheduleUpdated:
		return settings.NotifyScheduleUpdated
	}
	return true
}
