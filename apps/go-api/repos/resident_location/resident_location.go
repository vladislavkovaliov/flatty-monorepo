package residentlocation

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	residentlocationdomain "flatty-budget/go-api/domains/resident_location"
)

type PgxRepository struct {
	pool *pgxpool.Pool
}

func NewPgxRepository(pool *pgxpool.Pool) *PgxRepository {
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
