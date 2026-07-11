package resident_location

import (
	"context"
)

type Repository interface {
	Count(ctx context.Context) (int, error)
	List(ctx context.Context, limit, offset int) ([]*ResidentLocation, error)
	Create(ctx context.Context, input *ResidentLocationInput) (*ResidentLocation, error)
	Update(ctx context.Context, id int64, input *ResidentLocationInput) (*ResidentLocation, error)
	Delete(ctx context.Context, id int64) (int64, error)
}
