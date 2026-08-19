package expense_stats

import (
	expensestatsdomain "flatty-budget/go-api/domains/expense_stats"

	"github.com/jackc/pgx/v5"
)

func scanMonthlyTotal(row pgx.CollectableRow) (*expensestatsdomain.ExpenseMonthlyTotal, error) {
	var residentLocationID int64
	var month, year int
	var totalSpent float64
	if err := row.Scan(&residentLocationID, &month, &year, &totalSpent); err != nil {
		return nil, err
	}
	return expensestatsdomain.NewExpenseMonthlyTotal(residentLocationID, month, year, totalSpent), nil
}

func scanMonthlyAverage(row pgx.CollectableRow) (*expensestatsdomain.ExpenseMonthlyAverage, error) {
	var residentLocationID int64
	var month, year, expenseCount int
	var averageAmount float64
	if err := row.Scan(&residentLocationID, &month, &year, &averageAmount, &expenseCount); err != nil {
		return nil, err
	}
	return expensestatsdomain.NewExpenseMonthlyAverage(residentLocationID, month, year, averageAmount, expenseCount), nil
}
