package expense_stats

import "context"

type MonthlyTotalRepository interface {
	List(ctx context.Context, month, year *int) ([]*ExpenseMonthlyTotal, error)
	UpsertTotal(ctx context.Context, month, year int, totalSpent float64) error
}

type MonthlyAverageRepository interface {
	List(ctx context.Context, month, year *int) ([]*ExpenseMonthlyAverage, error)
	UpsertAverage(ctx context.Context, month, year int, averageAmount float64, expenseCount int) error
}
