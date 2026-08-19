package category

import (
	"time"

	"github.com/jackc/pgx/v5"

	categorydomain "flatty-budget/go-api/domains/category"
)

// scanCategory scans a single row with columns
// (id, name, description, created_at, updated_at) into a Category.
func scanCategory(row pgx.CollectableRow) (*categorydomain.Category, error) {
	var id int64
	var name string
	var description string
	var createdAt time.Time
	var updatedAt time.Time

	if err := row.Scan(&id, &name, &description, &createdAt, &updatedAt); err != nil {
		return nil, err
	}

	return categorydomain.NewCategory(id, name, description, createdAt, updatedAt), nil
}
