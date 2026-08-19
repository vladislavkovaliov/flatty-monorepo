package residentlocation

import (
	"time"

	residentlocationdomain "flatty-budget/go-api/domains/resident_location"

	"github.com/jackc/pgx/v5"
)

func scanResidentLocation(row pgx.CollectableRow) (*residentlocationdomain.ResidentLocation, error) {
	var id int64
	var userID string
	var country string
	var city string
	var postalCode string
	var street string
	var house string
	var apartment string
	var createdAt time.Time
	var updatedAt time.Time

	err := row.Scan(&id, &userID, &country, &city, &postalCode, &street, &house, &apartment, &createdAt, &updatedAt)
	if err != nil {
		return nil, err
	}

	return residentlocationdomain.NewResidentLocation(
		id, userID, country, city, postalCode, street, house, apartment, createdAt, updatedAt,
	), nil
}
