package expenses

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"

	expensedomain "flatty-budget/go-api/domains/expenses"
)

type pgxPool interface {
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
}

type PgxRepository struct {
	pool pgxPool
}

func NewPgxRepository(pool pgxPool) *PgxRepository {
	return &PgxRepository{
		pool: pool,
	}
}

func (r *PgxRepository) Count(ctx context.Context, residentLocationID int64, userID string) (int, error) {
	var count int

	err := r.pool.QueryRow(ctx, sqlExpensesCount, residentLocationID, userID).Scan(&count)

	return count, err
}

func (r *PgxRepository) List(ctx context.Context, residentLocationID int64, userID string, limit, offset int) ([]*expensedomain.ExpenseWithCategory, error) {
	rows, err := r.pool.Query(ctx, sqlExpensesList, residentLocationID, userID, limit, offset)

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	return pgx.CollectRows(rows, scanExpenseWithCategory)
}

func (r *PgxRepository) GetByID(ctx context.Context, id int64) (*expensedomain.Expense, error) {
	rows, err := r.pool.Query(ctx, sqlGetExpenseByID, id)

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	expense, err := pgx.CollectOneRow(rows, scanExpense)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("expense with id %d not found: %w", id, pgx.ErrNoRows)
		}

		return nil, err
	}

	return expense, nil
}

func (r *PgxRepository) Create(ctx context.Context, input *expensedomain.ExpenseInput, userID string) (*expensedomain.Expense, error) {
	rows, err := r.pool.Query(ctx, sqlExpenseCreate,
		userID,
		input.ResidentLocationID(),
		input.CategoryID(),
		input.Amount(),
		input.Month(),
		input.Year(),
		input.Description(),
	)

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	return pgx.CollectOneRow(rows, scanExpense)
}

func (r *PgxRepository) Update(ctx context.Context, id int64, input *expensedomain.ExpenseInput, userID string) (*expensedomain.Expense, error) {
	rows, err := r.pool.Query(ctx, sqlExpenseUpdate,
		input.ResidentLocationID(),
		input.CategoryID(),
		input.Amount(),
		input.Month(),
		input.Year(),
		input.Description(),
		id,
		userID,
	)

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	expense, err := pgx.CollectOneRow(rows, scanExpense)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("expense with id %d not found: %w", id, pgx.ErrNoRows)
		}

		return nil, err
	}

	return expense, nil
}

func (r *PgxRepository) Delete(ctx context.Context, id int64, userID string) (int64, error) {
	var returningID int64

	err := r.pool.QueryRow(ctx, sqlExpenseDelete, id, userID).Scan(&returningID)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return -1, fmt.Errorf("expense with id %d not found: %w", id, pgx.ErrNoRows)
		}

		return -1, err
	}

	return returningID, nil
}

func (r *PgxRepository) GetYearsAndMonths(ctx context.Context, residentLocationID int64, userID string) ([]*expensedomain.YearAndMonth, error) {
	rows, err := r.pool.Query(ctx, sqlGetYearsAndMonths, residentLocationID, userID)

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	return pgx.CollectRows(rows, scanYearAndMonth)
}

func (r *PgxRepository) GetExpensesByYearMonth(ctx context.Context, residentLocationID, year, month int64, userID string) ([]*expensedomain.ExpenseWithCategory, error) {
	rows, err := r.pool.Query(ctx, sqlGetExpenseByYearMonth, residentLocationID, userID, year, month)

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	return pgx.CollectRows(rows, scanExpenseWithCategory)
}
