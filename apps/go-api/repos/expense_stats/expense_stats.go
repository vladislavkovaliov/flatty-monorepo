package expense_stats

import (
	"context"

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
	query := sqlListMonthlyTotalsBase
	args := []any{residentLocationID, userID}
	query, args = appendMonthYearFilters(query, args, month, year)
	query += sqlOrderByMonthYearDesc

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return pgx.CollectRows(rows, scanMonthlyTotal)
}

func (r *PgxMonthlyTotalRepository) UpsertTotal(ctx context.Context, residentLocationID int64, month, year int, totalSpent float64) error {
	_, err := r.pool.Exec(ctx, sqlUpsertMonthlyTotal, residentLocationID, month, year, totalSpent)
	return err
}

type PgxMonthlyAverageRepository struct {
	pool pgxPool
}

func NewPgxMonthlyAverageRepository(pool pgxPool) *PgxMonthlyAverageRepository {
	return &PgxMonthlyAverageRepository{pool: pool}
}

func (r *PgxMonthlyAverageRepository) List(ctx context.Context, residentLocationID int64, userID string, month, year *int) ([]*expensestatsdomain.ExpenseMonthlyAverage, error) {
	query := sqlListMonthlyAveragesBase
	args := []any{residentLocationID, userID}
	query, args = appendMonthYearFilters(query, args, month, year)
	query += sqlOrderByMonthYearDesc

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return pgx.CollectRows(rows, scanMonthlyAverage)
}

func (r *PgxMonthlyAverageRepository) UpsertAverage(ctx context.Context, residentLocationID int64, month, year int, averageAmount float64, expenseCount int) error {
	_, err := r.pool.Exec(ctx, sqlUpsertMonthlyAverage, residentLocationID, month, year, averageAmount, expenseCount)
	return err
}
