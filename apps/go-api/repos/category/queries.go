package category

// sqlCategoryColumns is shared by List and the RETURNING clauses of
// Create and Update.
const sqlCategoryColumns = "id, name, description, created_at, updated_at"

const (
	sqlCategoriesCount = `SELECT COUNT(*) FROM categories`

	sqlCategoriesList = `SELECT ` + sqlCategoryColumns + `
		FROM categories LIMIT $1 OFFSET $2`

	sqlCategoryCreate = `INSERT INTO categories (name, description)
		VALUES ($1, $2)
		RETURNING ` + sqlCategoryColumns

	sqlCategoryUpdate = `UPDATE categories
		SET
			name = $1,
			description = $2,
			updated_at = NOW()
		WHERE id = $3
		RETURNING ` + sqlCategoryColumns

	sqlCategoryDelete = `DELETE FROM categories
		WHERE id = $1
		RETURNING id`
)
