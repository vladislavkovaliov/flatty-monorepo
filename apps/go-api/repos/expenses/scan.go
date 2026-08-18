package expenses

import (
	categorydomain "flatty-budget/go-api/domains/category"
	expensedomain "flatty-budget/go-api/domains/expenses"
	"time"

	"github.com/jackc/pgx/v5"
)

func scanExpenseWithCategory(row pgx.CollectableRow) (*expensedomain.ExpenseWithCategory, error) {
	var (
		id                 int64
		residentLocationID int64
		categoryID         int64
		amount             float64
		description        string
		month              int
		year               int
		createdAt          time.Time
		updatedAt          time.Time
		userID             string

		categoryName        string
		categoryDescription string
		categoryCreatedAt   time.Time
		categoryUpdatedAt   time.Time
	)

	if err := row.Scan(
		&id, &residentLocationID, &categoryID, &amount, &month, &year,
		&createdAt, &updatedAt, &description, &userID,
		&categoryName, &categoryDescription, &categoryCreatedAt, &categoryUpdatedAt,
	); err != nil {
		return nil, err
	}

	return expensedomain.NewExpenseWithCategory(
		id, residentLocationID, categoryID, amount, description, month, year,
		createdAt, updatedAt, userID,
		*categorydomain.NewCategory(
			categoryID, categoryName, categoryDescription, categoryCreatedAt, categoryUpdatedAt,
		),
	), nil
}

func scanExpense(row pgx.CollectableRow) (*expensedomain.Expense, error) {
	var (
		id                 int64
		residentLocationID int64
		categoryID         int64
		amount             float64
		description        string
		month              int
		year               int
		createdAt          time.Time
		updatedAt          time.Time
		userID             string
	)

	if err := row.Scan(
		&id, &residentLocationID, &categoryID, &amount, &month, &year,
		&createdAt, &updatedAt, &description, &userID,
	); err != nil {
		return nil, err
	}

	return expensedomain.NewExpense(id, residentLocationID, categoryID, amount, description, month, year, createdAt, updatedAt, userID), nil
}

func scanYearAndMonth(row pgx.CollectableRow) (*expensedomain.YearAndMonth, error) {
	var (
		year     int64
		month    int64
		expenses int64
	)

	if err := row.Scan(&year, &month, &expenses); err != nil {
		return nil, err
	}

	return expensedomain.NewYearAndMonth(year, month, expenses), nil
}
