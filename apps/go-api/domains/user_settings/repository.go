package user_settings

import "context"

type UserSettingsRepository interface {
	GetByUserID(ctx context.Context, userID string) (*UserSettings, error)
	Upsert(ctx context.Context, userID string, input *UserSettingsInput) (*UserSettings, error)
}