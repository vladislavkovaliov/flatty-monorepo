package user

import (
	"context"
	userdomain "flatty-budget/go-api/domains/user"
	"time"

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
	rows, err := r.pool.Query(ctx, `
		SELECT "id", name, email, "emailVerified", image, "createdAt", "updatedAt"
		FROM "user"
		LIMIT $1 OFFSET $2
	`, limit, offset)

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var users []*userdomain.User

	for rows.Next() {
		var id string
		var name string
		var email string
		var emailVerified bool
		var image *string
		var createdAt time.Time
		var updatedAt time.Time

		if err := rows.Scan(&id, &name, &email, &emailVerified, &image, &createdAt, &updatedAt); err != nil {
			if err == pgx.ErrNoRows {
				return nil, nil
			}
		}

		users = append(users,
			userdomain.NewUser(id, name, email, emailVerified, image, createdAt, updatedAt),
		)
	}

	return users, err
}

func (p *PgxRepository) GetUserByID(ctx context.Context, userId string) (*userdomain.User, error) {
	var id string
	var name string
	var email string
	var emailVerified bool
	var image *string
	var createdAt time.Time
	var updatedAt time.Time

	err := p.pool.QueryRow(ctx, `
		SELECT "id", name, email, "emailVerified", image, "createdAt", "updatedAt"
        FROM "user"
        WHERE "id" = $1
	`, userId).Scan(&id, &name, &email, &emailVerified, &image, &createdAt, &updatedAt)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	return userdomain.NewUser(id, name, email, emailVerified, image, createdAt, updatedAt), nil
}
