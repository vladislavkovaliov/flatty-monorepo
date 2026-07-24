package user_settings

import (
	"context"
	"time"

	user_settings_domain "flatty-budget/go-api/domains/user_settings"
)

type Service struct {
	repo user_settings_domain.UserSettingsRepository
}

func NewService(repo user_settings_domain.UserSettingsRepository) *Service {
	return &Service{
		repo: repo,
	}
}

func (s *Service) Get(ctx context.Context, userID string) (*user_settings_domain.UserSettings, error) {
	settings, err := s.repo.GetByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	if settings == nil {
		now := time.Now()
		return user_settings_domain.NewUserSettings(userID, "en", "system", "UTC+0", "YYYY-MM-DD", now, now), nil
	}

	return settings, nil
}

func (s *Service) Update(ctx context.Context, userID string, input *user_settings_domain.UserSettingsInput) (*user_settings_domain.UserSettings, error) {
	return s.repo.Upsert(ctx, userID, input)
}
