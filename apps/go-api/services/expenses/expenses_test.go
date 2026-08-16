package expenses

import (
	"context"
	"errors"
	"testing"
	"time"

	categorydomain "flatty-budget/go-api/domains/category"
	"flatty-budget/go-api/domains/expenses"
	kafkaclient "flatty-budget/go-api/internal/kafka"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// mockRepo implements expensedomain.Repository.
type mockRepo struct {
	mock.Mock
}

func (m *mockRepo) Count(ctx context.Context, residentLocationID int64, userID string) (int, error) {
	args := m.Called(ctx, residentLocationID, userID)
	return args.Int(0), args.Error(1)
}

func (m *mockRepo) List(ctx context.Context, residentLocationID int64, userID string, limit, offset int) ([]*expenses.ExpenseWithCategory, error) {
	args := m.Called(ctx, residentLocationID, userID, limit, offset)
	return args.Get(0).([]*expenses.ExpenseWithCategory), args.Error(1)
}

func (m *mockRepo) GetByID(ctx context.Context, id int64) (*expenses.Expense, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(*expenses.Expense), args.Error(1)
}

func (m *mockRepo) Create(ctx context.Context, input *expenses.ExpenseInput, userID string) (*expenses.Expense, error) {
	args := m.Called(ctx, input, userID)
	return args.Get(0).(*expenses.Expense), args.Error(1)
}

func (m *mockRepo) Update(ctx context.Context, id int64, input *expenses.ExpenseInput, userID string) (*expenses.Expense, error) {
	args := m.Called(ctx, id, input, userID)
	return args.Get(0).(*expenses.Expense), args.Error(1)
}

func (m *mockRepo) Delete(ctx context.Context, id int64, userID string) (int64, error) {
	args := m.Called(ctx, id, userID)
	return args.Get(0).(int64), args.Error(1)
}

func (m *mockRepo) GetYearsAndMonths(ctx context.Context, residentLocationID int64, userID string) ([]*expenses.YearAndMonth, error) {
	args := m.Called(ctx, residentLocationID, userID)
	return args.Get(0).([]*expenses.YearAndMonth), args.Error(1)
}

func (m *mockRepo) GetExpensesByYearMonth(ctx context.Context, residentLocationID, year, month int64, userID string) ([]*expenses.ExpenseWithCategory, error) {
	args := m.Called(ctx, residentLocationID, year, month, userID)
	return args.Get(0).([]*expenses.ExpenseWithCategory), args.Error(1)
}

// mockProducer implements kafkaclient.ProducerInterface.
type mockProducer struct {
	mock.Mock
}

func (m *mockProducer) PublishEvent(ctx context.Context, event kafkaclient.ExpenseEvent) error {
	args := m.Called(ctx, event)
	return args.Error(0)
}

func TestService_Count(t *testing.T) {
	t.Parallel()

	type countCase struct {
		name    string
		repoErr error
		repoRes int
		want    int
		wantErr string
	}

	cases := []countCase{
		{
			name:    "success",
			repoRes: 10,
			want:    10,
		},
		{
			name:    "repo error",
			repoErr: errors.New("db error"),
			wantErr: "db error",
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			repo := new(mockRepo)
			svc := New(repo, nil)

			residentLocationID := int64(1)
			userID := "test-user-id"

			repo.On("Count", mock.Anything, residentLocationID, userID).Return(tc.repoRes, tc.repoErr)

			got, err := svc.Count(context.Background(), residentLocationID, userID)

			if tc.wantErr != "" {
				assert.Error(t, err)
				assert.Equal(t, tc.wantErr, err.Error())
				assert.Equal(t, 0, got)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.want, got)
			}

			repo.AssertExpectations(t)
		})
	}
}

func TestService_List(t *testing.T) {
	t.Parallel()

	now := time.Now()
	cat1 := categorydomain.NewCategory(1, "Food", "", now, now)
	cat2 := categorydomain.NewCategory(2, "Transport", "", now, now)
	exp1 := expenses.NewExpenseWithCategory(1, 1, 1, 100.0, "", 1, 2024, now, now, "123456", *cat1)
	exp2 := expenses.NewExpenseWithCategory(2, 1, 2, 200.0, "", 1, 2024, now, now, "123456", *cat2)

	type listCase struct {
		name          string
		limit, offset int

		listRepoErr error
		listRepoRes []*expenses.ExpenseWithCategory

		countRepoErr error
		countRepoRes int

		want      []*expenses.ExpenseWithCategory
		wantTotal int
		wantErr   string
	}

	cases := []listCase{
		{
			name:   "success",
			limit:  10,
			offset: 0,

			listRepoRes:  []*expenses.ExpenseWithCategory{exp1, exp2},
			countRepoRes: 2,

			want:      []*expenses.ExpenseWithCategory{exp1, exp2},
			wantTotal: 2,
		},
		{
			name:   "list repo error",
			limit:  10,
			offset: 0,

			listRepoErr: errors.New("list error"),

			wantErr: "list error",
		},
		{
			name:   "count repo error",
			limit:  10,
			offset: 0,

			listRepoRes:  []*expenses.ExpenseWithCategory{exp1},
			countRepoErr: errors.New("count error"),

			wantErr: "count error",
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			repo := new(mockRepo)
			svc := New(repo, nil)

			residentLocationID := int64(1)
			userID := "test-user-id"

			repo.On("List", mock.Anything, residentLocationID, userID, tc.limit, tc.offset).Return(tc.listRepoRes, tc.listRepoErr)

			if tc.listRepoErr == nil {
				repo.On("Count", mock.Anything, residentLocationID, userID).Return(tc.countRepoRes, tc.countRepoErr)
			}

			got, total, err := svc.List(context.Background(), residentLocationID, userID, tc.limit, tc.offset)

			if tc.wantErr != "" {
				assert.Error(t, err)
				assert.Equal(t, tc.wantErr, err.Error())
				assert.Nil(t, got)
				assert.Equal(t, 0, total)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.want, got)
				assert.Equal(t, tc.wantTotal, total)
			}

			repo.AssertExpectations(t)
		})
	}
}

func TestService_Create(t *testing.T) {
	t.Parallel()

	now := time.Now()
	input := expenses.NewExpenseInput(1, 2, 150.0, "", 3, 2024)
	expected := expenses.NewExpense(1, 1, 2, 150.0, "", 3, 2024, now, now, "123456")

	type createCase struct {
		name        string
		input       *expenses.ExpenseInput
		repoRes     *expenses.Expense
		repoErr     error
		producerNil bool
		pubErr      error
		want        *expenses.Expense
		wantErr     string
	}

	cases := []createCase{
		{
			name:        "success without producer",
			input:       input,
			repoRes:     expected,
			producerNil: true,
			want:        expected,
		},
		{
			name:    "success with producer",
			input:   input,
			repoRes: expected,
			want:    expected,
		},
		{
			name:    "success with producer publish error (logged, not returned)",
			input:   input,
			repoRes: expected,
			pubErr:  errors.New("kafka down"),
			want:    expected,
		},
		{
			name:        "repo error",
			input:       input,
			repoErr:     errors.New("create error"),
			producerNil: true,
			wantErr:     "create error",
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			repo := new(mockRepo)
			var producer kafkaclient.ProducerInterface

			if tc.producerNil {
				producer = nil
			} else {
				mockProd := new(mockProducer)
				expectedEvent := kafkaclient.ExpenseEvent{
					Action:             "created",
					ID:                 expected.ID(),
					ResidentLocationID: expected.ResidentLocationID(),
					Month:              expected.Month(),
					Year:               expected.Year(),
					Amount:             expected.Amount(),
				}
				mockProd.On("PublishEvent", mock.Anything, expectedEvent).Return(tc.pubErr)
				producer = mockProd
				defer mockProd.AssertExpectations(t)
			}

			svc := New(repo, producer)
			userID := "test-user-id"
			repo.On("Create", mock.Anything, tc.input, userID).Return(tc.repoRes, tc.repoErr)

			got, err := svc.Create(context.Background(), tc.input, userID)

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

func TestService_Update(t *testing.T) {
	t.Parallel()

	now := time.Now()
	prevExpense := expenses.NewExpense(1, 1, 2, 100.0, "", 3, 2024, now, now, "123456")
	input := expenses.NewExpenseInput(1, 2, 200.0, "", 3, 2024)
	updatedExpense := expenses.NewExpense(1, 1, 2, 200.0, "", 3, 2024, now, now, "123456")

	type updateCase struct {
		name        string
		id          int64
		input       *expenses.ExpenseInput
		getRepoRes  *expenses.Expense
		getRepoErr  error
		updRepoRes  *expenses.Expense
		updRepoErr  error
		producerNil bool
		pubErr      error
		want        *expenses.Expense
		wantErr     string
	}

	cases := []updateCase{
		{
			name:        "success without producer",
			id:          1,
			input:       input,
			getRepoRes:  prevExpense,
			updRepoRes:  updatedExpense,
			producerNil: true,
			want:        updatedExpense,
		},
		{
			name:       "success with producer",
			id:         1,
			input:      input,
			getRepoRes: prevExpense,
			updRepoRes: updatedExpense,
			want:       updatedExpense,
		},
		{
			name:       "success with producer publish error (logged, not returned)",
			id:         1,
			input:      input,
			getRepoRes: prevExpense,
			updRepoRes: updatedExpense,
			pubErr:     errors.New("kafka down"),
			want:       updatedExpense,
		},
		{
			name:       "get repo error",
			id:         1,
			input:      input,
			getRepoErr: errors.New("not found"),
			wantErr:    "not found",
		},
		{
			name:       "update repo error",
			id:         1,
			input:      input,
			getRepoRes: prevExpense,
			updRepoErr: errors.New("update error"),
			wantErr:    "update error",
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			repo := new(mockRepo)
			var producer kafkaclient.ProducerInterface

			repo.On("GetByID", mock.Anything, tc.id).Return(tc.getRepoRes, tc.getRepoErr)
			if tc.getRepoErr == nil {
				userID := "test-user-id"
				repo.On("Update", mock.Anything, tc.id, tc.input, userID).Return(tc.updRepoRes, tc.updRepoErr)
			}

			if !tc.producerNil && tc.getRepoErr == nil && tc.updRepoErr == nil {
				mockProd := new(mockProducer)
				expectedEvent := kafkaclient.ExpenseEvent{
					Action:             "updated",
					ID:                 updatedExpense.ID(),
					ResidentLocationID: updatedExpense.ResidentLocationID(),
					Month:              updatedExpense.Month(),
					Year:               updatedExpense.Year(),
					Amount:             updatedExpense.Amount(),
					PrevAmount:         prevExpense.Amount(),
				}
				mockProd.On("PublishEvent", mock.Anything, expectedEvent).Return(tc.pubErr)
				producer = mockProd
				defer mockProd.AssertExpectations(t)
			}

			svc := New(repo, producer)

			userID := "test-user-id"
			got, err := svc.Update(context.Background(), tc.id, tc.input, userID)

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

func TestService_Delete(t *testing.T) {
	t.Parallel()

	now := time.Now()
	prevExpense := expenses.NewExpense(1, 1, 2, 100.0, "", 3, 2024, now, now, "123456")

	type deleteCase struct {
		name        string
		id          int64
		getRepoRes  *expenses.Expense
		getRepoErr  error
		delRepoRes  int64
		delRepoErr  error
		producerNil bool
		pubErr      error
		want        int64
		wantErr     string
	}

	cases := []deleteCase{
		{
			name:        "success without producer",
			id:          1,
			getRepoRes:  prevExpense,
			delRepoRes:  1,
			producerNil: true,
			want:        1,
		},
		{
			name:       "success with producer",
			id:         1,
			getRepoRes: prevExpense,
			delRepoRes: 1,
			want:       1,
		},
		{
			name:       "success with producer publish error (logged, not returned)",
			id:         1,
			getRepoRes: prevExpense,
			delRepoRes: 1,
			pubErr:     errors.New("kafka down"),
			want:       1,
		},
		{
			name:       "get repo error",
			id:         1,
			getRepoErr: errors.New("not found"),
			wantErr:    "not found",
		},
		{
			name:       "delete repo error",
			id:         1,
			getRepoRes: prevExpense,
			delRepoErr: errors.New("delete error"),
			wantErr:    "delete error",
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			repo := new(mockRepo)
			var producer kafkaclient.ProducerInterface

			repo.On("GetByID", mock.Anything, tc.id).Return(tc.getRepoRes, tc.getRepoErr)
			if tc.getRepoErr == nil {
				userID := "test-user-id"
				repo.On("Delete", mock.Anything, tc.id, userID).Return(tc.delRepoRes, tc.delRepoErr)
			}

			if !tc.producerNil && tc.getRepoErr == nil && tc.delRepoErr == nil {
				mockProd := new(mockProducer)
				expectedEvent := kafkaclient.ExpenseEvent{
					Action:             "deleted",
					ID:                 prevExpense.ID(),
					ResidentLocationID: prevExpense.ResidentLocationID(),
					Month:              prevExpense.Month(),
					Year:               prevExpense.Year(),
					Amount:             prevExpense.Amount(),
				}
				mockProd.On("PublishEvent", mock.Anything, expectedEvent).Return(tc.pubErr)
				producer = mockProd
				defer mockProd.AssertExpectations(t)
			}

			svc := New(repo, producer)

			userID := "test-user-id"
			got, err := svc.Delete(context.Background(), tc.id, userID)

			if tc.wantErr != "" {
				assert.Error(t, err)
				assert.Equal(t, tc.wantErr, err.Error())
				assert.Equal(t, int64(0), got)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.want, got)
			}

			repo.AssertExpectations(t)
		})
	}
}
