package user_settings

import (
	"context"
	"errors"

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
	rows, err := r.pool.Query(ctx, sqlGetUserSettingsByUserID, userID)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	settings, err := pgx.CollectOneRow(rows, scanUserSettings)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}

	return settings, err
}

func (r *PgxRepository) Upsert(ctx context.Context, userID string, input *user_settings_domain.UserSettingsInput) (*user_settings_domain.UserSettings, error) {
	rows, err := r.pool.Query(ctx, sqlUpsertUserSettings,
		userID, input.Language(), input.Theme(), input.Timezone(), input.DateFormat(),
	)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	return pgx.CollectOneRow(rows, scanUserSettings)
}
