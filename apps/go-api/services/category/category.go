package category

import (
	"context"

	categorydomain "flatty-budget/go-api/domains/category"
)

type Service struct {
	repo categorydomain.Repository
}

func New(repo categorydomain.Repository) *Service {
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

func (s *Service) List(ctx context.Context, limit, offset int) ([]*categorydomain.Category, int, error) {
	items, err := s.repo.List(ctx, limit, offset)

	if err != nil {
		return nil, 0, err
	}

	total, err := s.repo.Count(ctx)

	if err != nil {
		return nil, 0, err
	}

	return items, total, err
}

func (s *Service) Create(ctx context.Context, input *categorydomain.CategoryInput) (*categorydomain.Category, error) {
	return s.repo.Create(ctx, input)
}

func (s *Service) Update(ctx context.Context, id int64, input *categorydomain.CategoryInput) (*categorydomain.Category, error) {
	return s.repo.Update(ctx, id, input)
}

func (s *Service) Delete(ctx context.Context, id int64) (int64, error) {
	return s.repo.Delete(ctx, id)
}
