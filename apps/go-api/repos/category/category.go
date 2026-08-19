package category

import (
	"context"
	"errors"
	"fmt"

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

	err := r.pool.QueryRow(ctx, sqlCategoriesCount).Scan(&count)

	return count, err
}

func (r *PgxRepository) List(ctx context.Context, limit, offset int) ([]*categorydomain.Category, error) {
	rows, err := r.pool.Query(ctx, sqlCategoriesList, limit, offset)

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	return pgx.CollectRows(rows, scanCategory)
}

func (r *PgxRepository) Create(ctx context.Context, input *categorydomain.CategoryInput) (*categorydomain.Category, error) {
	rows, err := r.pool.Query(ctx, sqlCategoryCreate, input.Name(), input.Description())

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	return pgx.CollectOneRow(rows, scanCategory)
}

func (r *PgxRepository) Update(ctx context.Context, id int64, input *categorydomain.CategoryInput) (*categorydomain.Category, error) {
	rows, err := r.pool.Query(ctx, sqlCategoryUpdate, input.Name(), input.Description(), id)

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	category, err := pgx.CollectOneRow(rows, scanCategory)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("category with id %d not found: %w", id, pgx.ErrNoRows)
		}

		return nil, err
	}

	return category, nil
}

func (r *PgxRepository) Delete(ctx context.Context, id int64) (int64, error) {
	var returningId int64

	err := r.pool.QueryRow(ctx, sqlCategoryDelete, id).Scan(&returningId)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return -1, fmt.Errorf("category with id %d not found: %w", id, pgx.ErrNoRows)
		}

		return -1, err
	}

	return returningId, err
}
