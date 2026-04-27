package service

import (
	"context"
	"errors"

	"github.com/manatsanan0209/Vibe-Voyage_Backend/internal/domain"
	"gorm.io/gorm"
)

var (
	validThemes      = map[string]bool{"light": true, "dark": true, "system": true}
	validLanguages   = map[string]bool{"th": true, "en": true}
	validDateFormats = map[string]bool{"DD/MM/YYYY": true, "MM/DD/YYYY": true, "YYYY-MM-DD": true}
	validTimeFormats = map[string]bool{"12h": true, "24h": true}
)

type userSettingsService struct {
	repo domain.UserSettingsRepository
}

func NewUserSettingsService(repo domain.UserSettingsRepository) domain.UserSettingsService {
	return &userSettingsService{repo: repo}
}

func (s *userSettingsService) GetSettings(ctx context.Context, userID uint) (*domain.UserSettings, error) {
	settings, err := s.repo.GetByUserID(ctx, userID)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return &domain.UserSettings{
			UserID:     userID,
			Theme:      "system",
			Language:   "en",
			DateFormat: "DD/MM/YYYY",
			TimeFormat: "24h",
		}, nil
	}
	return settings, err
}

func (s *userSettingsService) UpdateSettings(ctx context.Context, userID uint, input domain.UpdateUserSettingsInput) (*domain.UserSettings, error) {
	if err := validateInput(input); err != nil {
		return nil, err
	}

	settings, err := s.GetSettings(ctx, userID)
	if err != nil {
		return nil, err
	}

	if input.Theme != nil {
		settings.Theme = *input.Theme
	}
	if input.Language != nil {
		settings.Language = *input.Language
	}
	if input.DateFormat != nil {
		settings.DateFormat = *input.DateFormat
	}
	if input.TimeFormat != nil {
		settings.TimeFormat = *input.TimeFormat
	}
	if input.NotifyRoomInvite != nil {
		settings.NotifyRoomInvite = *input.NotifyRoomInvite
	}
	if input.NotifyMemberJoined != nil {
		settings.NotifyMemberJoined = *input.NotifyMemberJoined
	}
	if input.NotifyMemberLeft != nil {
		settings.NotifyMemberLeft = *input.NotifyMemberLeft
	}
	if input.NotifyTripCreated != nil {
		settings.NotifyTripCreated = *input.NotifyTripCreated
	}
	if input.NotifyLifestyleAnalyzed != nil {
		settings.NotifyLifestyleAnalyzed = *input.NotifyLifestyleAnalyzed
	}
	if input.NotifyScheduleUpdated != nil {
		settings.NotifyScheduleUpdated = *input.NotifyScheduleUpdated
	}

	if err := s.repo.Upsert(ctx, settings); err != nil {
		return nil, err
	}
	return settings, nil
}

func validateInput(input domain.UpdateUserSettingsInput) error {
	if input.Theme != nil && !validThemes[*input.Theme] {
		return errors.New("invalid theme: must be light, dark, or system")
	}
	if input.Language != nil && !validLanguages[*input.Language] {
		return errors.New("invalid language: must be th or en")
	}
	if input.DateFormat != nil && !validDateFormats[*input.DateFormat] {
		return errors.New("invalid date_format: must be DD/MM/YYYY, MM/DD/YYYY, or YYYY-MM-DD")
	}
	if input.TimeFormat != nil && !validTimeFormats[*input.TimeFormat] {
		return errors.New("invalid time_format: must be 12h or 24h")
	}
	return nil
}
