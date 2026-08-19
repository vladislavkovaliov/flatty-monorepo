package user_settings

const sqlUserSettingsColumns = `user_id, language, theme, timezone, date_format, created_at, updated_at`

const (
	sqlGetUserSettingsByUserID = `SELECT ` + sqlUserSettingsColumns + ` FROM user_settings WHERE user_id = $1`
	sqlUpsertUserSettings      = `INSERT INTO user_settings (user_id, language, theme, timezone, date_format) VALUES ($1, $2, $3, $4, $5) ON CONFLICT (user_id) DO UPDATE SET language = EXCLUDED.language, theme = EXCLUDED.theme, timezone = EXCLUDED.timezone, date_format = EXCLUDED.date_format, updated_at = NOW() RETURNING ` + sqlUserSettingsColumns
)
