package expenses

const (
	sqlExpensesCount = `
		SELECT COUNT(*) FROM expenses
		WHERE resident_location_id = $1 AND user_id = $2
	`
	sqlExpensesList = `
		SELECT 
			e.id, 
			e.resident_location_id, 
			e.category_id, 
			e.amount, 
			e.month,
			e.year, 
			e.created_at, 
			e.updated_at, 
			e.description, 
			e.user_id, 
			c.name as category_name, 
			c.description as category_description,
			c.created_at as category_created_at,
			c.updated_at as category_updated_at
		FROM expenses e
		LEFT JOIN categories c ON e.category_id = c.id
		WHERE resident_location_id = $1 AND user_id = $2
		ORDER BY id 
		LIMIT $3 OFFSET $4
	`
	sqlGetExpenseByID = `
		SELECT id, resident_location_id, category_id, amount, month, year, created_at, updated_at, description, user_id
		FROM expenses
		WHERE id = $1
	`
	sqlExpenseCreate = `
		INSERT INTO expenses (user_id, resident_location_id, category_id, amount, month, year, description)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id, resident_location_id, category_id, amount, month, year, created_at, updated_at, description, user_id
	`
	sqlExpenseUpdate = `
		UPDATE expenses
		SET
			resident_location_id = $1,
			category_id = $2,
			amount = $3,
			month = $4,
			year = $5,
			updated_at = NOW(),
			description = $6
		WHERE id = $7 AND user_id = $8
		RETURNING id, resident_location_id, category_id, amount, month, year, created_at, updated_at, description, user_id
	`
	sqlExpenseDelete = `
		DELETE FROM expenses
		WHERE id = $1 AND user_id = $2
		RETURNING id
	`
	sqlGetYearsAndMonths = `
		SELECT year, month, COUNT(*) as "expcenses" FROM expenses 
		WHERE resident_location_id = $1 AND user_id = $2
		GROUP BY year, month
		ORDER BY year ASC, month ASC
		LIMIT 100;
	`
	sqlGetExpenseByYearMonth = `
		SELECT 
			e.id, 
			e.resident_location_id, 
			e.category_id, 
			e.amount, 
			e.month,
			e.year, 
			e.created_at, 
			e.updated_at, 
			e.description, 
			e.user_id, 
			c.name as category_name, 
			c.description as category_description,
			c.created_at as category_created_at,
			c.updated_at as category_updated_at
		FROM expenses e
		LEFT JOIN categories c ON e.category_id = c.id
		WHERE 
			e.resident_location_id = $1 
			AND e.user_id = $2
			AND e.year = $3 
			AND e.month = $4
		ORDER BY e.year ASC, e.month ASC
	`
)
