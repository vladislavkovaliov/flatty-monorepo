package expenses

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"

	expensedomain "flatty-budget/go-api/domains/expenses"
)

// pgxPool is a minimal interface matching the Query and QueryRow methods of *pgxpool.Pool.
// It exists to enable unit testing with mock implementations.
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

func (r *PgxRepository) Count(ctx context.Context) (int, error) {
	var count int

	err := r.pool.QueryRow(ctx, `
		SELECT COUNT(*) FROM expenses
	`).Scan(&count)

	return count, err
}

func (r *PgxRepository) List(ctx context.Context, limit, offset int) ([]*expensedomain.Expense, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, resident_location_id, category_id, amount, month, year, created_at, updated_at, description
		FROM expenses
		ORDER BY id 
		LIMIT $1 OFFSET $2
	`, limit, offset)

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var expenses []*expensedomain.Expense

	for rows.Next() {
		var id int64
		var residentLocationID int64
		var categoryID int64
		var amount float64
		var description string
		var month int
		var year int
		var createdAt time.Time
		var updatedAt time.Time

		if err := rows.Scan(&id, &residentLocationID, &categoryID, &amount, &month, &year, &createdAt, &updatedAt, &description); err != nil {
			return nil, err
		}

		expenses = append(expenses,
			expensedomain.NewExpense(id, residentLocationID, categoryID, amount, description, month, year, createdAt, updatedAt),
		)
	}

	return expenses, nil
}

func (r *PgxRepository) GetByID(ctx context.Context, id int64) (*expensedomain.Expense, error) {
	var expenseID int64
	var residentLocationID int64
	var categoryID int64
	var amount float64
	var description string
	var month int
	var year int
	var createdAt time.Time
	var updatedAt time.Time

	err := r.pool.QueryRow(ctx, `
		SELECT id, resident_location_id, category_id, amount, month, year, created_at, updated_at, description
		FROM expenses
		WHERE id = $1
	`, id).Scan(&expenseID, &residentLocationID, &categoryID, &amount, &month, &year, &createdAt, &updatedAt, &description)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("expense with id %d not found: %w", id, pgx.ErrNoRows)
		}
		return nil, err
	}

	return expensedomain.NewExpense(expenseID, residentLocationID, categoryID, amount, description, month, year, createdAt, updatedAt), nil
}

func (r *PgxRepository) Create(ctx context.Context, input *expensedomain.ExpenseInput) (*expensedomain.Expense, error) {
	var id int64
	var residentLocationID int64
	var categoryID int64
	var amount float64
	var description string
	var month int
	var year int
	var createdAt time.Time
	var updatedAt time.Time

	err := r.pool.QueryRow(ctx, `
		INSERT INTO expenses (resident_location_id, category_id, amount, month, year, description)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, resident_location_id, category_id, amount, month, year, created_at, updated_at, description
	`,
		input.ResidentLocationID(),
		input.CategoryID(),
		input.Amount(),
		input.Month(),
		input.Year(),
		input.Description(),
	).Scan(
		&id,
		&residentLocationID,
		&categoryID,
		&amount,
		&month,
		&year,
		&createdAt,
		&updatedAt,
		&description,
	)

	if err != nil {
		return nil, err
	}

	return expensedomain.NewExpense(id, residentLocationID, categoryID, amount, description, month, year, createdAt, updatedAt), nil
}

func (r *PgxRepository) Update(ctx context.Context, id int64, input *expensedomain.ExpenseInput) (*expensedomain.Expense, error) {
	var returningID int64
	var residentLocationID int64
	var categoryID int64
	var amount float64
	var description string
	var month int
	var year int
	var createdAt time.Time
	var updatedAt time.Time

	err := r.pool.QueryRow(ctx, `
		UPDATE expenses
		SET
			resident_location_id = $1,
			category_id = $2,
			amount = $3,
			month = $4,
			year = $5,
			updated_at = NOW(),
			description = $6
		WHERE id = $7
		RETURNING id, resident_location_id, category_id, amount, month, year, created_at, updated_at, description
	`,
		input.ResidentLocationID(),
		input.CategoryID(),
		input.Amount(),
		input.Month(),
		input.Year(),
		input.Description(),
		id,
	).Scan(
		&returningID,
		&residentLocationID,
		&categoryID,
		&amount,
		&month,
		&year,
		&createdAt,
		&updatedAt,
		&description,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("expense with id %d not found: %w", id, pgx.ErrNoRows)
		}

		return nil, err
	}

	return expensedomain.NewExpense(returningID, residentLocationID, categoryID, amount, description, month, year, createdAt, updatedAt), nil
}

func (r *PgxRepository) Delete(ctx context.Context, id int64) (int64, error) {
	var returningID int64

	err := r.pool.QueryRow(ctx, `
		DELETE FROM expenses
		WHERE id = $1
		RETURNING id
	`, id).Scan(&returningID)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return -1, fmt.Errorf("expense with id %d not found: %w", id, pgx.ErrNoRows)
		}

		return -1, err
	}

	return returningID, nil
}
