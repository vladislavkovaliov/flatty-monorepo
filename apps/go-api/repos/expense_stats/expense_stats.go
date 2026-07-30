package expense_stats

import (
	"context"
	"fmt"

	expensestatsdomain "flatty-budget/go-api/domains/expense_stats"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

// pgxPool is a minimal interface matching the Query, QueryRow, and Exec methods of *pgxpool.Pool.
// It exists to enable unit testing with mock implementations.
type pgxPool interface {
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
	Exec(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error)
}

type PgxMonthlyTotalRepository struct {
	pool pgxPool
}

func NewPgxMonthlyTotalRepository(pool pgxPool) *PgxMonthlyTotalRepository {
	return &PgxMonthlyTotalRepository{pool: pool}
}

func (r *PgxMonthlyTotalRepository) List(ctx context.Context, residentLocationID int64, userID string, month, year *int) ([]*expensestatsdomain.ExpenseMonthlyTotal, error) {
	query := `
		SELECT t.resident_location_id, t.month, t.year, t.total_spent
		FROM expense_monthly_totals t
		WHERE t.resident_location_id = $1
			AND EXISTS (
				SELECT 1 FROM resident_locations rl
				WHERE rl.id = t.resident_location_id AND rl.user_id = $2
			)`
	args := []interface{}{residentLocationID, userID}
	argIdx := 3

	if month != nil {
		query += fmt.Sprintf(` AND t.month = $%d`, argIdx)
		args = append(args, *month)
		argIdx++
	}
	if year != nil {
		query += fmt.Sprintf(` AND t.year = $%d`, argIdx)
		args = append(args, *year)
		argIdx++
	}
	query += ` ORDER BY t.year DESC, t.month DESC`

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var totals []*expensestatsdomain.ExpenseMonthlyTotal
	for rows.Next() {
		var rl int64
		var m, y int
		var ts float64
		if err := rows.Scan(&rl, &m, &y, &ts); err != nil {
			return nil, err
		}
		totals = append(totals, expensestatsdomain.NewExpenseMonthlyTotal(rl, m, y, ts))
	}
	return totals, nil
}

func (r *PgxMonthlyTotalRepository) UpsertTotal(ctx context.Context, residentLocationID int64, month, year int, totalSpent float64) error {
	_, err := r.pool.Exec(ctx, `
		INSERT INTO expense_monthly_totals (resident_location_id, month, year, total_spent, updated_at)
		VALUES ($1, $2, $3, $4, NOW())
		ON CONFLICT (resident_location_id, month, year) DO UPDATE SET
			total_spent = $4,
			updated_at = NOW()
	`, residentLocationID, month, year, totalSpent)
	return err
}

type PgxMonthlyAverageRepository struct {
	pool pgxPool
}

func NewPgxMonthlyAverageRepository(pool pgxPool) *PgxMonthlyAverageRepository {
	return &PgxMonthlyAverageRepository{pool: pool}
}

func (r *PgxMonthlyAverageRepository) List(ctx context.Context, residentLocationID int64, userID string, month, year *int) ([]*expensestatsdomain.ExpenseMonthlyAverage, error) {
	query := `
		SELECT a.resident_location_id, a.month, a.year, a.average_amount, a.expense_count
		FROM expense_monthly_averages a
		WHERE a.resident_location_id = $1
			AND EXISTS (
				SELECT 1 FROM resident_locations rl
				WHERE rl.id = a.resident_location_id AND rl.user_id = $2
			)`
	args := []interface{}{residentLocationID, userID}
	argIdx := 3

	if month != nil {
		query += fmt.Sprintf(` AND a.month = $%d`, argIdx)
		args = append(args, *month)
		argIdx++
	}
	if year != nil {
		query += fmt.Sprintf(` AND a.year = $%d`, argIdx)
		args = append(args, *year)
		argIdx++
	}
	query += ` ORDER BY a.year DESC, a.month DESC`

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var averages []*expensestatsdomain.ExpenseMonthlyAverage
	for rows.Next() {
		var rl int64
		var m, y, cnt int
		var avg float64
		if err := rows.Scan(&rl, &m, &y, &avg, &cnt); err != nil {
			return nil, err
		}
		averages = append(averages, expensestatsdomain.NewExpenseMonthlyAverage(rl, m, y, avg, cnt))
	}
	return averages, nil
}

func (r *PgxMonthlyAverageRepository) UpsertAverage(ctx context.Context, residentLocationID int64, month, year int, averageAmount float64, expenseCount int) error {
	_, err := r.pool.Exec(ctx, `
		INSERT INTO expense_monthly_averages (resident_location_id, month, year, average_amount, expense_count, updated_at)
		VALUES ($1, $2, $3, $4, $5, NOW())
		ON CONFLICT (resident_location_id, month, year) DO UPDATE SET
			average_amount = $4,
			expense_count = $5,
			updated_at = NOW()
	`, residentLocationID, month, year, averageAmount, expenseCount)
	return err
}
