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

func (s *Service) Count(ctx context.Context) (int, error) {
	count, err := s.repo.Count(ctx)

	if err != nil {
		return 0, err
	}

	return count, err
}
