package expenses

import "context"

type Repository interface {
	Count(ctx context.Context) (int, error)
	List(ctx context.Context, limit, offset int) ([]*Expense, error)
	GetByID(ctx context.Context, id int64) (*Expense, error)
	Create(ctx context.Context, input *ExpenseInput) (*Expense, error)
	Update(ctx context.Context, id int64, input *ExpenseInput) (*Expense, error)
	Delete(ctx context.Context, id int64) (int64, error)
}
