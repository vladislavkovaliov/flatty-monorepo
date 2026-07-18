package residentlocation

import (
	"context"
	residentlocationdomain "flatty-budget/go-api/domains/resident_location"
)

type Service struct {
	repo residentlocationdomain.Repository
}

func New(repo residentlocationdomain.Repository) *Service {
	return &Service{
		repo: repo,
	}
}

func (s *Service) Count(ctx context.Context, userID string) (int, error) {
	count, err := s.repo.Count(ctx, userID)

	if err != nil {
		return 0, err
	}

	return count, err
}

func (s *Service) List(ctx context.Context, limit, offset int, userID string) ([]*residentlocationdomain.ResidentLocation, int, error) {
	items, err := s.repo.List(ctx, limit, offset, userID)

	if err != nil {
		return nil, 0, err
	}

	total, err := s.repo.Count(ctx, userID)

	if err != nil {
		return nil, 0, err
	}

	return items, total, err
}

func (s *Service) Create(ctx context.Context, input *residentlocationdomain.ResidentLocationInput, userID string) (*residentlocationdomain.ResidentLocation, error) {
	return s.repo.Create(ctx, input, userID)
}

func (s *Service) Update(ctx context.Context, id int64, input *residentlocationdomain.ResidentLocationInput, userID string) (*residentlocationdomain.ResidentLocation, error) {
	return s.repo.Update(ctx, id, input, userID)
}

func (s *Service) Delete(ctx context.Context, id int64, userID string) (int64, error) {
	return s.repo.Delete(ctx, id, userID)
}
