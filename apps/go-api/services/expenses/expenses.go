package expenses

import (
	"context"
	"fmt"

	expensedomain "flatty-budget/go-api/domains/expenses"
)

type Service struct {
	repo expensedomain.Repository
}

func New(repo expensedomain.Repository) *Service {
	return &Service{
		repo: repo,
	}
}

func (s *Service) Count(ctx context.Context) (int, error) {
	count, err := s.repo.Count(ctx)

	if err != nil {
		return 0, err
	}

	return count, nil
}

func (s *Service) List(ctx context.Context, limit, offset int) ([]*expensedomain.Expense, int, error) {
	items, err := s.repo.List(ctx, limit, offset)

	if err != nil {
		return nil, 0, err
	}

	total, err := s.repo.Count(ctx)
	fmt.Println(total)

	if err != nil {
		return nil, 0, err
	}

	return items, total, nil
}

func (s *Service) Create(ctx context.Context, input *expensedomain.ExpenseInput) (*expensedomain.Expense, error) {
	return s.repo.Create(ctx, input)
}

func (s *Service) Update(ctx context.Context, id int64, input *expensedomain.ExpenseInput) (*expensedomain.Expense, error) {
	return s.repo.Update(ctx, id, input)
}

func (s *Service) Delete(ctx context.Context, id int64) (int64, error) {
	return s.repo.Delete(ctx, id)
}
