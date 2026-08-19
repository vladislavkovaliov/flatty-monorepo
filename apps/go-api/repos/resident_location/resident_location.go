package residentlocation

import (
	"context"
	"errors"
	"fmt"

	residentlocationdomain "flatty-budget/go-api/domains/resident_location"

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

func (r *PgxRepository) Count(ctx context.Context, userID string) (int, error) {
	var count int

	err := r.pool.QueryRow(ctx, sqlResidentLocationsCount, userID).Scan(&count)

	return count, err
}

func (r *PgxRepository) List(ctx context.Context, limit, offset int, userID string) ([]*residentlocationdomain.ResidentLocation, error) {
	rows, err := r.pool.Query(ctx, sqlResidentLocationsList, userID, limit, offset)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	return pgx.CollectRows(rows, scanResidentLocation)
}

func (r *PgxRepository) Create(ctx context.Context, input *residentlocationdomain.ResidentLocationInput, userID string) (*residentlocationdomain.ResidentLocation, error) {
	rows, err := r.pool.Query(ctx, sqlResidentLocationCreate,
		userID, input.Country(), input.City(), input.PostalCode(), input.Street(), input.House(), input.Apartment(),
	)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	return pgx.CollectOneRow(rows, scanResidentLocation)
}

func (r *PgxRepository) Update(ctx context.Context, id int64, input *residentlocationdomain.ResidentLocationInput, userID string) (*residentlocationdomain.ResidentLocation, error) {
	rows, err := r.pool.Query(ctx, sqlResidentLocationUpdate,
		input.Country(), input.City(), input.PostalCode(), input.Street(), input.House(), input.Apartment(), id, userID,
	)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	location, err := pgx.CollectOneRow(rows, scanResidentLocation)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, fmt.Errorf("resident location with id %d not found: %w", id, pgx.ErrNoRows)
	}

	return location, err
}

func (r *PgxRepository) Delete(ctx context.Context, id int64, userID string) (int64, error) {
	var returningID int64

	err := r.pool.QueryRow(ctx, sqlResidentLocationDelete, id, userID).Scan(&returningID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return -1, fmt.Errorf("resident location with id %d not found: %w", id, pgx.ErrNoRows)
		}

		return -1, err
	}

	return returningID, nil
}
