package applications

import (
	"context"

	applicationsdomain "flatty-budget/go-api/domains/applications"
)

type Service struct {
	repo applicationsdomain.Repository
}

func New(repo applicationsdomain.Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) ListAll(ctx context.Context) ([]*applicationsdomain.Application, error) {
	return s.repo.ListAll(ctx)
}

func (s *Service) ListByEnv(ctx context.Context, env string) ([]*applicationsdomain.Application, error) {
	return s.repo.ListByEnv(ctx, env)
}

func (s *Service) Create(ctx context.Context, input *applicationsdomain.ApplicationInput) (*applicationsdomain.Application, error) {
	return s.repo.Create(ctx, input)
}

func (s *Service) Update(ctx context.Context, id int64, input *applicationsdomain.ApplicationInput) (*applicationsdomain.Application, error) {
	return s.repo.Update(ctx, id, input)
}

func (s *Service) Delete(ctx context.Context, id int64) (int64, error) {
	return s.repo.Delete(ctx, id)
}
