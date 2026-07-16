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

func (r *PgxMonthlyTotalRepository) List(ctx context.Context, month, year *int) ([]*expensestatsdomain.ExpenseMonthlyTotal, error) {
	query := `SELECT month, year, total_spent FROM expense_monthly_totals WHERE TRUE`
	var args []interface{}
	argIdx := 1

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
	query += ` ORDER BY year DESC, month DESC`

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var totals []*expensestatsdomain.ExpenseMonthlyTotal
	for rows.Next() {
		var m, y int
		var ts float64
		if err := rows.Scan(&m, &y, &ts); err != nil {
			return nil, err
		}
		totals = append(totals, expensestatsdomain.NewExpenseMonthlyTotal(m, y, ts))
	}
	return totals, nil
}

func (r *PgxMonthlyTotalRepository) UpsertTotal(ctx context.Context, month, year int, totalSpent float64) error {
	_, err := r.pool.Exec(ctx, `
		INSERT INTO expense_monthly_totals (month, year, total_spent, updated_at)
		VALUES ($1, $2, $3, NOW())
		ON CONFLICT (month, year) DO UPDATE SET
			total_spent = $3,
			updated_at = NOW()
	`, month, year, totalSpent)
	return err
}

type PgxMonthlyAverageRepository struct {
	pool pgxPool
}

func NewPgxMonthlyAverageRepository(pool pgxPool) *PgxMonthlyAverageRepository {
	return &PgxMonthlyAverageRepository{pool: pool}
}

func (r *PgxMonthlyAverageRepository) List(ctx context.Context, month, year *int) ([]*expensestatsdomain.ExpenseMonthlyAverage, error) {
	query := `SELECT month, year, average_amount, expense_count FROM expense_monthly_averages WHERE TRUE`
	var args []interface{}
	argIdx := 1

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
	query += ` ORDER BY year DESC, month DESC`

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var averages []*expensestatsdomain.ExpenseMonthlyAverage
	for rows.Next() {
		var m, y, cnt int
		var avg float64
		if err := rows.Scan(&m, &y, &avg, &cnt); err != nil {
			return nil, err
		}
		averages = append(averages, expensestatsdomain.NewExpenseMonthlyAverage(m, y, avg, cnt))
	}
	return averages, nil
}

func (r *PgxMonthlyAverageRepository) UpsertAverage(ctx context.Context, month, year int, averageAmount float64, expenseCount int) error {
	_, err := r.pool.Exec(ctx, `
		INSERT INTO expense_monthly_averages (month, year, average_amount, expense_count, updated_at)
		VALUES ($1, $2, $3, $4, NOW())
		ON CONFLICT (month, year) DO UPDATE SET
			average_amount = $3,
			expense_count = $4,
			updated_at = NOW()
	`, month, year, averageAmount, expenseCount)
	return err
}
