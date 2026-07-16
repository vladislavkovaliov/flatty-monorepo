package category

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"

	categorydomain "flatty-budget/go-api/domains/category"
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

func (r *PgxRepository) Count(ctx context.Context) (int, error) {
	var count int

	err := r.pool.QueryRow(ctx, `
		SELECT COUNT(*) FROM categories
	`).Scan(&count)

	return count, err
}

func (r *PgxRepository) List(ctx context.Context, limit, offset int) ([]*categorydomain.Category, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, name, description, created_at, updated_at
		FROM categories LIMIT $1 OFFSET $2
	`, limit, offset)

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var categories []*categorydomain.Category

	for rows.Next() {
		var id int64
		var name string
		var description string
		var created_at time.Time
		var updated_at time.Time

		if err := rows.Scan(&id, &name, &description, &created_at, &updated_at); err != nil {
			return nil, err
		}

		categories = append(categories,
			categorydomain.NewCategory(id, name, description, created_at, updated_at),
		)
	}

	return categories, err
}

func (r *PgxRepository) Create(ctx context.Context, input *categorydomain.CategoryInput) (*categorydomain.Category, error) {
	var id int64
	var name string
	var description string
	var createdAt time.Time
	var updatedAt time.Time

	err := r.pool.QueryRow(ctx, `
		INSERT INTO categories (name, description)
		VALUES ($1, $2)
		RETURNING id, name, description, created_at, updated_at
	`, input.Name(), input.Description()).Scan(
		&id,
		&name,
		&description,
		&createdAt,
		&updatedAt,
	)

	if err != nil {
		return nil, err
	}

	return categorydomain.NewCategory(id, name, description, createdAt, updatedAt), nil
}

func (r *PgxRepository) Update(ctx context.Context, id int64, input *categorydomain.CategoryInput) (*categorydomain.Category, error) {
	var returningId int64
	var name string
	var description string
	var createdAt time.Time
	var updatedAt time.Time

	err := r.pool.QueryRow(ctx, `
		UPDATE categories
		SET
			name = $1,
			description = $2,
			updated_at = NOW()
		WHERE id = $3
		RETURNING id, name, description, created_at, updated_at
	`, input.Name(), input.Description(), id).Scan(
		&returningId,
		&name,
		&description,
		&createdAt,
		&updatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("category with id %d not found: %w", id, pgx.ErrNoRows)
		}

		return nil, err
	}

	return categorydomain.NewCategory(returningId, name, description, createdAt, updatedAt), nil
}

func (r *PgxRepository) Delete(ctx context.Context, id int64) (int64, error) {
	var returningId int64

	err := r.pool.QueryRow(ctx, `
		DELETE FROM categories
		WHERE id = $1
		RETURNING id
	`, id).Scan(&returningId)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return -1, fmt.Errorf("category with id %d not found: %w", id, pgx.ErrNoRows)
		}

		return -1, err
	}

	return returningId, err
}
