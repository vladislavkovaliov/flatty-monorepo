package resident_location

import "context"

type Repository interface {
	Count(ctx context.Context) (int, error)
	List(ctx context.Context, limit, offset int) ([]*ResidentLocation, error)
}
