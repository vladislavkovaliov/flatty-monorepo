package category

import "context"

type Repository interface {
	Count(ctx context.Context) (int, error)
	List(ctx context.Context, limit, offset int) ([]*Category, error)
	Create(ctx context.Context, input *CategoryInput) (*Category, error)
	Update(ctx context.Context, id int64, input *CategoryInput) (*Category, error)
	Delete(ctx context.Context, id int64) (int64, error)
}
