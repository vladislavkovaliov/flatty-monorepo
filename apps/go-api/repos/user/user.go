package user

import (
	"context"
	"errors"

	userdomain "flatty-budget/go-api/domains/user"

	"github.com/jackc/pgx/v5"
)

// pgxPool is a minimal interface matching the Query and QueryRow methods of *pgxpool.Pool.
// It exists to enable unit testing with mock implementations.
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

func (r *PgxRepository) List(ctx context.Context, limit, offset int) ([]*userdomain.User, error) {
	rows, err := r.pool.Query(ctx, sqlUsersList, limit, offset)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	return pgx.CollectRows(rows, scanUser)
}

func (r *PgxRepository) GetUserByID(ctx context.Context, userID string) (*userdomain.User, error) {
	rows, err := r.pool.Query(ctx, sqlGetUserByID, userID)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	user, err := pgx.CollectOneRow(rows, scanUser)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}

	return user, err
}
