package expense_stats

import (
	"context"

	expensestatsdomain "flatty-budget/go-api/domains/expense_stats"
)

type MonthlyTotalService struct {
	repo expensestatsdomain.MonthlyTotalRepository
}

func NewMonthlyTotalService(repo expensestatsdomain.MonthlyTotalRepository) *MonthlyTotalService {
	return &MonthlyTotalService{repo: repo}
}

func (s *MonthlyTotalService) List(ctx context.Context, month, year *int) ([]*expensestatsdomain.ExpenseMonthlyTotal, error) {
	return s.repo.List(ctx, month, year)
}

type MonthlyAverageService struct {
	repo expensestatsdomain.MonthlyAverageRepository
}

func NewMonthlyAverageService(repo expensestatsdomain.MonthlyAverageRepository) *MonthlyAverageService {
	return &MonthlyAverageService{repo: repo}
}

func (s *MonthlyAverageService) List(ctx context.Context, month, year *int) ([]*expensestatsdomain.ExpenseMonthlyAverage, error) {
	return s.repo.List(ctx, month, year)
}
