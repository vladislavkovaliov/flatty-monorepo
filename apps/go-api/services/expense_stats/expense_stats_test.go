package expense_stats

import (
	"context"
	"errors"
	"testing"

	"flatty-budget/go-api/domains/expense_stats"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// mockMonthlyTotalRepo implements expensestatsdomain.MonthlyTotalRepository.
type mockMonthlyTotalRepo struct {
	mock.Mock
}

func (m *mockMonthlyTotalRepo) List(ctx context.Context, month, year *int) ([]*expense_stats.ExpenseMonthlyTotal, error) {
	args := m.Called(ctx, month, year)
	return args.Get(0).([]*expense_stats.ExpenseMonthlyTotal), args.Error(1)
}

func (m *mockMonthlyTotalRepo) UpsertTotal(ctx context.Context, month, year int, totalSpent float64) error {
	args := m.Called(ctx, month, year, totalSpent)
	return args.Error(0)
}

// mockMonthlyAverageRepo implements expensestatsdomain.MonthlyAverageRepository.
type mockMonthlyAverageRepo struct {
	mock.Mock
}

func (m *mockMonthlyAverageRepo) List(ctx context.Context, month, year *int) ([]*expense_stats.ExpenseMonthlyAverage, error) {
	args := m.Called(ctx, month, year)
	return args.Get(0).([]*expense_stats.ExpenseMonthlyAverage), args.Error(1)
}

func (m *mockMonthlyAverageRepo) UpsertAverage(ctx context.Context, month, year int, averageAmount float64, expenseCount int) error {
	args := m.Called(ctx, month, year, averageAmount, expenseCount)
	return args.Error(0)
}

func intPtr(v int) *int {
	return &v
}

func TestMonthlyTotalService_List(t *testing.T) {
	t.Parallel()

	total1 := expense_stats.NewExpenseMonthlyTotal(1, 2024, 500.0)
	total2 := expense_stats.NewExpenseMonthlyTotal(2, 2024, 750.0)

	type listCase struct {
		name         string
		month, year  *int
		repoRes      []*expense_stats.ExpenseMonthlyTotal
		repoErr      error
		want         []*expense_stats.ExpenseMonthlyTotal
		wantErr      string
	}

	cases := []listCase{
		{
			name:    "success with no filters",
			month:   nil,
			year:    nil,
			repoRes: []*expense_stats.ExpenseMonthlyTotal{total1, total2},
			want:    []*expense_stats.ExpenseMonthlyTotal{total1, total2},
		},
		{
			name:    "success with filters",
			month:   intPtr(1),
			year:    intPtr(2024),
			repoRes: []*expense_stats.ExpenseMonthlyTotal{total1},
			want:    []*expense_stats.ExpenseMonthlyTotal{total1},
		},
		{
			name:    "empty result",
			month:   nil,
			year:    nil,
			repoRes: []*expense_stats.ExpenseMonthlyTotal{},
			want:    []*expense_stats.ExpenseMonthlyTotal{},
		},
		{
			name:   "repo error",
			month:  nil,
			year:   nil,
			repoErr: errors.New("db error"),
			wantErr: "db error",
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			repo := new(mockMonthlyTotalRepo)
			svc := NewMonthlyTotalService(repo)

			repo.On("List", mock.Anything, tc.month, tc.year).Return(tc.repoRes, tc.repoErr)

			got, err := svc.List(context.Background(), tc.month, tc.year)

			if tc.wantErr != "" {
				assert.Error(t, err)
				assert.Equal(t, tc.wantErr, err.Error())
				assert.Nil(t, got)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.want, got)
			}

			repo.AssertExpectations(t)
		})
	}
}

func TestMonthlyAverageService_List(t *testing.T) {
	t.Parallel()

	avg1 := expense_stats.NewExpenseMonthlyAverage(1, 2024, 250.0, 2)
	avg2 := expense_stats.NewExpenseMonthlyAverage(2, 2024, 375.0, 2)

	type listCase struct {
		name         string
		month, year  *int
		repoRes      []*expense_stats.ExpenseMonthlyAverage
		repoErr      error
		want         []*expense_stats.ExpenseMonthlyAverage
		wantErr      string
	}

	cases := []listCase{
		{
			name:    "success with no filters",
			month:   nil,
			year:    nil,
			repoRes: []*expense_stats.ExpenseMonthlyAverage{avg1, avg2},
			want:    []*expense_stats.ExpenseMonthlyAverage{avg1, avg2},
		},
		{
			name:    "success with filters",
			month:   intPtr(1),
			year:    intPtr(2024),
			repoRes: []*expense_stats.ExpenseMonthlyAverage{avg1},
			want:    []*expense_stats.ExpenseMonthlyAverage{avg1},
		},
		{
			name:    "empty result",
			month:   nil,
			year:    nil,
			repoRes: []*expense_stats.ExpenseMonthlyAverage{},
			want:    []*expense_stats.ExpenseMonthlyAverage{},
		},
		{
			name:   "repo error",
			month:  nil,
			year:   nil,
			repoErr: errors.New("db error"),
			wantErr: "db error",
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			repo := new(mockMonthlyAverageRepo)
			svc := NewMonthlyAverageService(repo)

			repo.On("List", mock.Anything, tc.month, tc.year).Return(tc.repoRes, tc.repoErr)

			got, err := svc.List(context.Background(), tc.month, tc.year)

			if tc.wantErr != "" {
				assert.Error(t, err)
				assert.Equal(t, tc.wantErr, err.Error())
				assert.Nil(t, got)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.want, got)
			}

			repo.AssertExpectations(t)
		})
	}
}
