package expense_stats

import "context"

type MonthlyTotalRepository interface {
	List(ctx context.Context, residentLocationID int64, userID string, month, year *int) ([]*ExpenseMonthlyTotal, error)
	UpsertTotal(ctx context.Context, residentLocationID int64, month, year int, totalSpent float64) error
}

type MonthlyAverageRepository interface {
	List(ctx context.Context, residentLocationID int64, userID string, month, year *int) ([]*ExpenseMonthlyAverage, error)
	UpsertAverage(ctx context.Context, residentLocationID int64, month, year int, averageAmount float64, expenseCount int) error
}
