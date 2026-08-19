package user

const sqlUserColumns = `"id", name, email, "emailVerified", image, "createdAt", "updatedAt"`

const (
	sqlUsersList   = `SELECT ` + sqlUserColumns + ` FROM "user" LIMIT $1 OFFSET $2`
	sqlGetUserByID = `SELECT ` + sqlUserColumns + ` FROM "user" WHERE "id" = $1`
)
