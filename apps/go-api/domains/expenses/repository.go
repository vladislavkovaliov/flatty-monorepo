package expenses

import "context"

type Repository interface {
	Count(ctx context.Context, residentLocationID int64, userID string) (int, error)
	List(ctx context.Context, residentLocationID int64, userID string, limit, offset int) ([]*ExpenseWithCategory, error)
	Create(ctx context.Context, input *ExpenseInput, userID string) (*Expense, error)
	Update(ctx context.Context, id int64, input *ExpenseInput, userID string) (*Expense, error)
	Delete(ctx context.Context, id int64, userID string) (int64, error)
	GetByID(ctx context.Context, id int64) (*Expense, error)
	GetYearsAndMonths(ctx context.Context, residentLocationID int64, userID string) ([]*YearAndMonth, error)
	GetExpensesByYearMonth(ctx context.Context, residentLocationID, year, month int64, userID string) ([]*ExpenseWithCategory, error)
}
