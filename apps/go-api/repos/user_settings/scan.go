package user_settings

import (
	"time"

	user_settings_domain "flatty-budget/go-api/domains/user_settings"

	"github.com/jackc/pgx/v5"
)

func scanUserSettings(row pgx.CollectableRow) (*user_settings_domain.UserSettings, error) {
	var userID string
	var language string
	var theme string
	var timezone string
	var dateFormat string
	var createdAt time.Time
	var updatedAt time.Time

	err := row.Scan(&userID, &language, &theme, &timezone, &dateFormat, &createdAt, &updatedAt)
	if err != nil {
		return nil, err
	}

	return user_settings_domain.NewUserSettings(userID, language, theme, timezone, dateFormat, createdAt, updatedAt), nil
}
