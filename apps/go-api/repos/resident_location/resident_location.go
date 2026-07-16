package residentlocation

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"

	residentlocationdomain "flatty-budget/go-api/domains/resident_location"
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
		SELECT COUNT(*) FROM resident_locations
	`).Scan(&count)

	return count, err
}

func (r *PgxRepository) List(ctx context.Context, limit, offset int) ([]*residentlocationdomain.ResidentLocation, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, country, city, postal_code, street, house, apartment, created_at, updated_at 
		FROM resident_locations LIMIT $1 OFFSET $2
	`, limit, offset)

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var residentLocations []*residentlocationdomain.ResidentLocation

	for rows.Next() {
		var id int64
		var country string
		var city string
		var postal_code string
		var street string
		var house string
		var apartment string
		var created_at time.Time
		var updated_at time.Time

		if err := rows.Scan(&id, &country, &city, &postal_code, &street, &house, &apartment, &created_at, &updated_at); err != nil {
			return nil, err
		}

		residentLocations = append(residentLocations,
			residentlocationdomain.NewResidentLocation(
				id, country, city, postal_code, street, house, apartment, created_at, updated_at,
			),
		)
	}

	return residentLocations, err
}

func (r *PgxRepository) Create(ctx context.Context, input *residentlocationdomain.ResidentLocationInput) (*residentlocationdomain.ResidentLocation, error) {
	var id int64
	var country string
	var city string
	var postalCode string
	var street string
	var house string
	var apartment string
	var createdAt time.Time
	var updatedAt time.Time

	err := r.pool.QueryRow(ctx, `
		INSERT INTO resident_locations (country, city, postal_code, street, house, apartment) 
		VALUES ($1, $2, $3, $4, $5, $6) 
		RETURNING id, country, city, postal_code, street, house, apartment, created_at, updated_at
	`, input.Country(), input.City(), input.PostalCode(), input.Street(), input.House(), input.Apartment()).Scan(
		&id,
		&country,
		&city,
		&postalCode,
		&street,
		&house,
		&apartment,
		&createdAt,
		&updatedAt,
	)

	if err != nil {
		return nil, err
	}

	return residentlocationdomain.NewResidentLocation(
		id,
		country,
		city,
		postalCode,
		street,
		house,
		apartment,
		createdAt,
		updatedAt,
	), nil
}

func (r *PgxRepository) Update(ctx context.Context, id int64, input *residentlocationdomain.ResidentLocationInput) (*residentlocationdomain.ResidentLocation, error) {
	var returningId int64
	var country string
	var city string
	var postalCode string
	var street string
	var house string
	var apartment string
	var createdAt time.Time
	var updatedAt time.Time

	err := r.pool.QueryRow(ctx, `
		UPDATE resident_locations
		SET
			country = $1,
			city = $2, 
			postal_code = $3, 
			street = $4, 
			house = $5, 
			apartment = $6,
			updated_at = NOW() 
		WHERE id = $7 
		RETURNING id, country, city, postal_code, street, house, apartment, created_at, updated_at
	`, input.Country(), input.City(), input.PostalCode(), input.Street(), input.House(), input.Apartment(), id).Scan(
		&returningId,
		&country,
		&city,
		&postalCode,
		&street,
		&house,
		&apartment,
		&createdAt,
		&updatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("resident location with id %d not found: %w", id, pgx.ErrNoRows)
		}

		return nil, err
	}

	return residentlocationdomain.NewResidentLocation(
		returningId,
		country,
		city,
		postalCode,
		street,
		house,
		apartment,
		createdAt,
		updatedAt,
	), nil
}

func (r *PgxRepository) Delete(ctx context.Context, id int64) (int64, error) {
	var returningId int64

	err := r.pool.QueryRow(ctx, `
		DELETE FROM resident_locations
		WHERE id = $1
		RETURNING id
	`, id).Scan(&returningId)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return -1, fmt.Errorf("resident location with id %d not found: %w", id, pgx.ErrNoRows)
		}

		return -1, err
	}

	return returningId, err
}
