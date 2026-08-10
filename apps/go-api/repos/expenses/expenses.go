package expenses

import (
	"context"
	"errors"
	"fmt"
	"time"

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

	err := r.pool.QueryRow(ctx, `
		SELECT COUNT(*) FROM expenses
		WHERE resident_location_id = $1 AND user_id = $2
	`, residentLocationID, userID).Scan(&count)

	return count, err
}

func (r *PgxRepository) List(ctx context.Context, residentLocationID int64, userID string, limit, offset int) ([]*expensedomain.Expense, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, resident_location_id, category_id, amount, month, year, created_at, updated_at, description, user_id
		FROM expenses
		WHERE resident_location_id = $1 AND user_id = $2
		ORDER BY id 
		LIMIT $3 OFFSET $4
	`, residentLocationID, userID, limit, offset)

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
		var _userID string

		if err := rows.Scan(&id, &residentLocationID, &categoryID, &amount, &month, &year, &createdAt, &updatedAt, &description, &_userID); err != nil {
			return nil, err
		}

		expenses = append(expenses,
			expensedomain.NewExpense(id, residentLocationID, categoryID, amount, description, month, year, createdAt, updatedAt, _userID),
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
	var _userID string

	err := r.pool.QueryRow(ctx, `
		SELECT id, resident_location_id, category_id, amount, month, year, created_at, updated_at, description, user_id
		FROM expenses
		WHERE id = $1
	`, id).Scan(&expenseID, &residentLocationID, &categoryID, &amount, &month, &year, &createdAt, &updatedAt, &description, &_userID)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("expense with id %d not found: %w", id, pgx.ErrNoRows)
		}
		return nil, err
	}

	return expensedomain.NewExpense(expenseID, residentLocationID, categoryID, amount, description, month, year, createdAt, updatedAt, _userID), nil
}

func (r *PgxRepository) Create(ctx context.Context, input *expensedomain.ExpenseInput, userID string) (*expensedomain.Expense, error) {
	var id int64
	var residentLocationID int64
	var categoryID int64
	var amount float64
	var description string
	var month int
	var year int
	var createdAt time.Time
	var updatedAt time.Time
	var _userID string

	err := r.pool.QueryRow(ctx, `
		INSERT INTO expenses (user_id, resident_location_id, category_id, amount, month, year, description)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id, resident_location_id, category_id, amount, month, year, created_at, updated_at, description, user_id
	`,
		userID,
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
		&_userID,
	)

	if err != nil {
		return nil, err
	}

	return expensedomain.NewExpense(id, residentLocationID, categoryID, amount, description, month, year, createdAt, updatedAt, _userID), nil
}

func (r *PgxRepository) Update(ctx context.Context, id int64, input *expensedomain.ExpenseInput, userID string) (*expensedomain.Expense, error) {
	var returningID int64
	var residentLocationID int64
	var categoryID int64
	var amount float64
	var description string
	var month int
	var year int
	var createdAt time.Time
	var updatedAt time.Time
	var _userID string

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
		WHERE id = $7 AND user_id = $8
		RETURNING id, resident_location_id, category_id, amount, month, year, created_at, updated_at, description, user_id
	`,
		input.ResidentLocationID(),
		input.CategoryID(),
		input.Amount(),
		input.Month(),
		input.Year(),
		input.Description(),
		id,
		userID,
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
		&_userID,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("expense with id %d not found: %w", id, pgx.ErrNoRows)
		}

		return nil, err
	}

	return expensedomain.NewExpense(returningID, residentLocationID, categoryID, amount, description, month, year, createdAt, updatedAt, _userID), nil
}

func (r *PgxRepository) Delete(ctx context.Context, id int64, userID string) (int64, error) {
	var returningID int64

	err := r.pool.QueryRow(ctx, `
		DELETE FROM expenses
		WHERE id = $1 AND user_id = $2
		RETURNING id
	`, id, userID).Scan(&returningID)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return -1, fmt.Errorf("expense with id %d not found: %w", id, pgx.ErrNoRows)
		}

		return -1, err
	}

	return returningID, nil
}
