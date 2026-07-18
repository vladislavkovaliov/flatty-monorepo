package resident_location

import (
	"context"
)

type Repository interface {
	Count(ctx context.Context, userID string) (int, error)
	List(ctx context.Context, limit, offset int, userID string) ([]*ResidentLocation, error)
	Create(ctx context.Context, input *ResidentLocationInput, userID string) (*ResidentLocation, error)
	Update(ctx context.Context, id int64, input *ResidentLocationInput, userID string) (*ResidentLocation, error)
	Delete(ctx context.Context, id int64, userID string) (int64, error)
}
