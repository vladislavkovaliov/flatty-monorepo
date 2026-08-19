package user

import (
	"time"

	userdomain "flatty-budget/go-api/domains/user"

	"github.com/jackc/pgx/v5"
)

func scanUser(row pgx.CollectableRow) (*userdomain.User, error) {
	var id string
	var name string
	var email string
	var emailVerified bool
	var image *string
	var createdAt time.Time
	var updatedAt time.Time

	err := row.Scan(&id, &name, &email, &emailVerified, &image, &createdAt, &updatedAt)
	if err != nil {
		return nil, err
	}

	return userdomain.NewUser(id, name, email, emailVerified, image, createdAt, updatedAt), nil
}
