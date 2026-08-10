package expenses

import "context"

type Repository interface {
	Count(ctx context.Context, residentLocationID int64, userID string) (int, error)
	List(ctx context.Context, residentLocationID int64, userID string, limit, offset int) ([]*Expense, error)
	GetByID(ctx context.Context, id int64) (*Expense, error)
	Create(ctx context.Context, input *ExpenseInput, userID string) (*Expense, error)
	Update(ctx context.Context, id int64, input *ExpenseInput, userID string) (*Expense, error)
	Delete(ctx context.Context, id int64, userID string) (int64, error)
}
