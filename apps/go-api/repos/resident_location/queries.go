package residentlocation

const sqlResidentLocationColumns = `id, user_id, country, city, postal_code, street, house, apartment, created_at, updated_at`

const (
	sqlResidentLocationsCount = `SELECT COUNT(*) FROM resident_locations WHERE user_id = $1`
	sqlResidentLocationsList  = `SELECT ` + sqlResidentLocationColumns + ` FROM resident_locations WHERE user_id = $1 LIMIT $2 OFFSET $3`
	sqlResidentLocationCreate = `INSERT INTO resident_locations (user_id, country, city, postal_code, street, house, apartment) VALUES ($1, $2, $3, $4, $5, $6, $7) RETURNING ` + sqlResidentLocationColumns
	sqlResidentLocationUpdate = `UPDATE resident_locations SET country = $1, city = $2, postal_code = $3, street = $4, house = $5, apartment = $6, updated_at = NOW() WHERE id = $7 AND user_id = $8 RETURNING ` + sqlResidentLocationColumns
	sqlResidentLocationDelete = `DELETE FROM resident_locations WHERE id = $1 AND user_id = $2 RETURNING id`
)
