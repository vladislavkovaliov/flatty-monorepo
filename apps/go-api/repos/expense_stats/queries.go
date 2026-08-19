package expense_stats

import "fmt"

const (
	// sqlListMonthlyTotalsBase selects from expense_monthly_totals, scoped by
	// resident location ($1) and ownership via user_id ($2). Optional month/year
	// filters are appended by appendMonthYearFilters.
	sqlListMonthlyTotalsBase = `
		SELECT t.resident_location_id, t.month, t.year, t.total_spent
		FROM expense_monthly_totals t
		WHERE t.resident_location_id = $1
			AND EXISTS (
				SELECT 1 FROM resident_locations rl
				WHERE rl.id = t.resident_location_id AND rl.user_id = $2
			)`

	// sqlListMonthlyAveragesBase is the averages counterpart of sqlListMonthlyTotalsBase.
	sqlListMonthlyAveragesBase = `
		SELECT a.resident_location_id, a.month, a.year, a.average_amount, a.expense_count
		FROM expense_monthly_averages a
		WHERE a.resident_location_id = $1
			AND EXISTS (
				SELECT 1 FROM resident_locations rl
				WHERE rl.id = a.resident_location_id AND rl.user_id = $2
			)`

	sqlOrderByMonthYearDesc = ` ORDER BY year DESC, month DESC`

	sqlUpsertMonthlyTotal = `
		INSERT INTO expense_monthly_totals (resident_location_id, month, year, total_spent, updated_at)
		VALUES ($1, $2, $3, $4, NOW())
		ON CONFLICT (resident_location_id, month, year) DO UPDATE SET
			total_spent = $4,
			updated_at = NOW()`

	sqlUpsertMonthlyAverage = `
		INSERT INTO expense_monthly_averages (resident_location_id, month, year, average_amount, expense_count, updated_at)
		VALUES ($1, $2, $3, $4, $5, NOW())
		ON CONFLICT (resident_location_id, month, year) DO UPDATE SET
			average_amount = $4,
			expense_count = $5,
			updated_at = NOW()`
)

// appendMonthYearFilters appends optional month/year predicates to a List query.
// Fragments use unqualified column names: unambiguous because resident_locations
// (the only other table in scope) has no month/year columns. The original
// argument order is preserved, new filter arguments appended.
func appendMonthYearFilters(query string, args []any, month, year *int) (string, []any) {
	argIdx := len(args) + 1
	if month != nil {
		query += fmt.Sprintf(` AND month = $%d`, argIdx)
		args = append(args, *month)
		argIdx++
	}
	if year != nil {
		query += fmt.Sprintf(` AND year = $%d`, argIdx)
		args = append(args, *year)
		argIdx++
	}
	return query, args
}
