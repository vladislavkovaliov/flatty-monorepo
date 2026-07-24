package user_settings

import (
	"context"
	"time"

	user_settings_domain "flatty-budget/go-api/domains/user_settings"

	"github.com/jackc/pgx/v5"
)

type pgxPool interface {
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
}

type PgxRepository struct {
	pool pgxPool
}

func NewPgxRepository(pool pgxPool) *PgxRepository {
	return &PgxRepository{
		pool: pool,
	}
}

func (r *PgxRepository) GetByUserID(ctx context.Context, userID string) (*user_settings_domain.UserSettings, error) {
	var userIDOut string
	var language string
	var theme string
	var timezone string
	var dateFormat string
	var createdAt time.Time
	var updatedAt time.Time

	err := r.pool.QueryRow(ctx, `
		SELECT user_id, language, theme, timezone, date_format, created_at, updated_at
		FROM user_settings
		WHERE user_id = $1
	`, userID).Scan(&userIDOut, &language, &theme, &timezone, &dateFormat, &createdAt, &updatedAt)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	return user_settings_domain.NewUserSettings(userIDOut, language, theme, timezone, dateFormat, createdAt, updatedAt), nil
}

func (r *PgxRepository) Upsert(ctx context.Context, userID string, input *user_settings_domain.UserSettingsInput) (*user_settings_domain.UserSettings, error) {
	var userIDOut string
	var language string
	var theme string
	var timezone string
	var dateFormat string
	var createdAt time.Time
	var updatedAt time.Time

	err := r.pool.QueryRow(ctx, `
		INSERT INTO user_settings (user_id, language, theme, timezone, date_format)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (user_id) DO UPDATE SET
			language = EXCLUDED.language,
			theme = EXCLUDED.theme,
			timezone = EXCLUDED.timezone,
			date_format = EXCLUDED.date_format,
			updated_at = NOW()
		RETURNING user_id, language, theme, timezone, date_format, created_at, updated_at
	`, userID, input.Language(), input.Theme(), input.Timezone(), input.DateFormat()).
		Scan(&userIDOut, &language, &theme, &timezone, &dateFormat, &createdAt, &updatedAt)

	if err != nil {
		return nil, err
	}

	return user_settings_domain.NewUserSettings(userIDOut, language, theme, timezone, dateFormat, createdAt, updatedAt), nil
}