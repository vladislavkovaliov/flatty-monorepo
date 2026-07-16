package expense_stats

import (
	"context"
	"errors"
	"reflect"
	"testing"

	expensestatsdomain "flatty-budget/go-api/domains/expense_stats"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// ---------------------------------------------------------------------------
// Mock types
// ---------------------------------------------------------------------------

// mockPgxPool implements pgxPool using testify.Mock.
type mockPgxPool struct {
	mock.Mock
}

func (m *mockPgxPool) Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error) {
	callArgs := m.Called(ctx, sql, args)
	rows, _ := callArgs.Get(0).(pgx.Rows)
	err, _ := callArgs.Get(1).(error)
	return rows, err
}

func (m *mockPgxPool) QueryRow(ctx context.Context, sql string, args ...any) pgx.Row {
	callArgs := m.Called(ctx, sql, args)
	row, _ := callArgs.Get(0).(pgx.Row)
	return row
}

func (m *mockPgxPool) Exec(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error) {
	callArgs := m.Called(ctx, sql, args)
	tag, _ := callArgs.Get(0).(pgconn.CommandTag)
	err, _ := callArgs.Get(1).(error)
	return tag, err
}

// mockRows implements pgx.Rows for testing.
type mockRows struct {
	rows    [][]any
	index   int
	scanErr error
}

func newMockRows(data [][]any) *mockRows {
	return &mockRows{rows: data, index: -1}
}

func (m *mockRows) Next() bool {
	m.index++
	return m.index < len(m.rows)
}

func (m *mockRows) Scan(dest ...any) error {
	if m.scanErr != nil {
		return m.scanErr
	}
	if m.index < 0 || m.index >= len(m.rows) {
		return errors.New("scan called without Next or out of bounds")
	}
	row := m.rows[m.index]
	for i, d := range dest {
		if i >= len(row) {
			break
		}
		v := reflect.ValueOf(d)
		if v.Kind() != reflect.Ptr {
			continue
		}
		srcVal := reflect.ValueOf(row[i])
		if srcVal.IsValid() {
			v.Elem().Set(srcVal)
		}
	}
	return nil
}

func (m *mockRows) Close() {}

func (m *mockRows) Err() error { return m.scanErr }

func (m *mockRows) CommandTag() pgconn.CommandTag { return pgconn.CommandTag{} }

func (m *mockRows) FieldDescriptions() []pgconn.FieldDescription { return nil }

func (m *mockRows) Values() ([]any, error) { return nil, nil }

func (m *mockRows) RawValues() [][]byte { return nil }

func (m *mockRows) Conn() *pgx.Conn { return nil }

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

func assertMonthlyTotalEqual(t *testing.T, want, got *expensestatsdomain.ExpenseMonthlyTotal) {
	t.Helper()
	assert.Equal(t, want.Month(), got.Month())
	assert.Equal(t, want.Year(), got.Year())
	assert.Equal(t, want.TotalSpent(), got.TotalSpent())
}

func assertMonthlyTotalSliceEqual(t *testing.T, want, got []*expensestatsdomain.ExpenseMonthlyTotal) {
	t.Helper()
	assert.Equal(t, len(want), len(got))
	for i := range want {
		assertMonthlyTotalEqual(t, want[i], got[i])
	}
}

func assertMonthlyAverageEqual(t *testing.T, want, got *expensestatsdomain.ExpenseMonthlyAverage) {
	t.Helper()
	assert.Equal(t, want.Month(), got.Month())
	assert.Equal(t, want.Year(), got.Year())
	assert.Equal(t, want.AverageAmount(), got.AverageAmount())
	assert.Equal(t, want.ExpenseCount(), got.ExpenseCount())
}

func assertMonthlyAverageSliceEqual(t *testing.T, want, got []*expensestatsdomain.ExpenseMonthlyAverage) {
	t.Helper()
	assert.Equal(t, len(want), len(got))
	for i := range want {
		assertMonthlyAverageEqual(t, want[i], got[i])
	}
}

// ---------------------------------------------------------------------------
// TestPgxMonthlyTotalRepository_List
// ---------------------------------------------------------------------------

func TestPgxMonthlyTotalRepository_List(t *testing.T) {
	t.Parallel()

	type listCase struct {
		name      string
		month     *int
		year      *int
		rows      *mockRows
		queryErr  error
		want      []*expensestatsdomain.ExpenseMonthlyTotal
		wantErr   string
	}

	cases := []listCase{
		{
			name:  "success_no_filters",
			month: nil,
			year:  nil,
			rows: newMockRows([][]any{
				{1, 2024, 1500.50},
				{2, 2024, 2000.00},
			}),
			queryErr: nil,
			want: []*expensestatsdomain.ExpenseMonthlyTotal{
				expensestatsdomain.NewExpenseMonthlyTotal(1, 2024, 1500.50),
				expensestatsdomain.NewExpenseMonthlyTotal(2, 2024, 2000.00),
			},
			wantErr: "",
		},
		{
			name:  "success_with_month_filter",
			month: intPtr(1),
			year:  nil,
			rows: newMockRows([][]any{
				{1, 2024, 1500.50},
			}),
			queryErr: nil,
			want: []*expensestatsdomain.ExpenseMonthlyTotal{
				expensestatsdomain.NewExpenseMonthlyTotal(1, 2024, 1500.50),
			},
			wantErr: "",
		},
		{
			name:  "success_with_year_filter",
			month: nil,
			year:  intPtr(2024),
			rows: newMockRows([][]any{
				{1, 2024, 1500.50},
				{2, 2024, 2000.00},
			}),
			queryErr: nil,
			want: []*expensestatsdomain.ExpenseMonthlyTotal{
				expensestatsdomain.NewExpenseMonthlyTotal(1, 2024, 1500.50),
				expensestatsdomain.NewExpenseMonthlyTotal(2, 2024, 2000.00),
			},
			wantErr: "",
		},
		{
			name:  "empty",
			month: nil,
			year:  nil,
			rows:  newMockRows(nil),
			want:  nil,
		},
		{
			name:     "query_error",
			month:    nil,
			year:     nil,
			rows:     nil,
			queryErr: errors.New("connection failed"),
			want:     nil,
			wantErr:  "connection failed",
		},
		{
			name:  "scan_error",
			month: nil,
			year:  nil,
			rows: func() *mockRows {
				r := newMockRows([][]any{
					{1, 2024, 1500.50},
				})
				r.scanErr = errors.New("scan failed")
				return r
			}(),
			queryErr: nil,
			want:     nil,
			wantErr:  "scan failed",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			pool := new(mockPgxPool)
			repo := NewPgxMonthlyTotalRepository(pool)

			ctx := context.Background()

			var rows pgx.Rows
			if tc.rows != nil {
				rows = tc.rows
			}

			expectedArgs := buildExpectedArgs(tc.month, tc.year)
			pool.On("Query", ctx, mock.AnythingOfType("string"),
				mock.MatchedBy(argsMatcher(expectedArgs)),
			).Return(rows, tc.queryErr)

			got, err := repo.List(ctx, tc.month, tc.year)

			if tc.wantErr != "" {
				assert.Error(t, err)
				assert.Equal(t, tc.wantErr, err.Error())
				assert.Nil(t, got)
			} else {
				assert.NoError(t, err)
				if tc.want == nil {
					assert.Nil(t, got)
				} else {
					assertMonthlyTotalSliceEqual(t, tc.want, got)
				}
			}

			pool.AssertExpectations(t)
		})
	}
}

// ---------------------------------------------------------------------------
// TestPgxMonthlyTotalRepository_UpsertTotal
// ---------------------------------------------------------------------------

func TestPgxMonthlyTotalRepository_UpsertTotal(t *testing.T) {
	t.Parallel()

	type upsertCase struct {
		name       string
		month      int
		year       int
		totalSpent float64
		execErr    error
		wantErr    string
	}

	cases := []upsertCase{
		{
			name:       "success",
			month:      1,
			year:       2024,
			totalSpent: 1500.50,
			execErr:    nil,
			wantErr:    "",
		},
		{
			name:       "exec_error",
			month:      1,
			year:       2024,
			totalSpent: 1500.50,
			execErr:    errors.New("upsert failed"),
			wantErr:    "upsert failed",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			pool := new(mockPgxPool)
			repo := NewPgxMonthlyTotalRepository(pool)

			ctx := context.Background()

			pool.On("Exec", ctx, mock.AnythingOfType("string"),
				[]any{tc.month, tc.year, tc.totalSpent},
			).Return(pgconn.CommandTag{}, tc.execErr)

			err := repo.UpsertTotal(ctx, tc.month, tc.year, tc.totalSpent)

			if tc.wantErr != "" {
				assert.Error(t, err)
				assert.Equal(t, tc.wantErr, err.Error())
			} else {
				assert.NoError(t, err)
			}

			pool.AssertExpectations(t)
		})
	}
}

// ---------------------------------------------------------------------------
// TestPgxMonthlyAverageRepository_List
// ---------------------------------------------------------------------------

func TestPgxMonthlyAverageRepository_List(t *testing.T) {
	t.Parallel()

	type listCase struct {
		name     string
		month    *int
		year     *int
		rows     *mockRows
		queryErr error
		want     []*expensestatsdomain.ExpenseMonthlyAverage
		wantErr  string
	}

	cases := []listCase{
		{
			name:  "success_no_filters",
			month: nil,
			year:  nil,
			rows: newMockRows([][]any{
				{1, 2024, 150.05, 10},
				{2, 2024, 200.00, 10},
			}),
			queryErr: nil,
			want: []*expensestatsdomain.ExpenseMonthlyAverage{
				expensestatsdomain.NewExpenseMonthlyAverage(1, 2024, 150.05, 10),
				expensestatsdomain.NewExpenseMonthlyAverage(2, 2024, 200.00, 10),
			},
			wantErr: "",
		},
		{
			name:  "success_with_month_filter",
			month: intPtr(1),
			year:  nil,
			rows: newMockRows([][]any{
				{1, 2024, 150.05, 10},
			}),
			queryErr: nil,
			want: []*expensestatsdomain.ExpenseMonthlyAverage{
				expensestatsdomain.NewExpenseMonthlyAverage(1, 2024, 150.05, 10),
			},
			wantErr: "",
		},
		{
			name:  "success_with_year_filter",
			month: nil,
			year:  intPtr(2024),
			rows: newMockRows([][]any{
				{1, 2024, 150.05, 10},
			}),
			queryErr: nil,
			want: []*expensestatsdomain.ExpenseMonthlyAverage{
				expensestatsdomain.NewExpenseMonthlyAverage(1, 2024, 150.05, 10),
			},
			wantErr: "",
		},
		{
			name:  "empty",
			month: nil,
			year:  nil,
			rows:  newMockRows(nil),
			want:  nil,
		},
		{
			name:     "query_error",
			month:    nil,
			year:     nil,
			rows:     nil,
			queryErr: errors.New("connection failed"),
			want:     nil,
			wantErr:  "connection failed",
		},
		{
			name:  "scan_error",
			month: nil,
			year:  nil,
			rows: func() *mockRows {
				r := newMockRows([][]any{
					{1, 2024, 150.05, 10},
				})
				r.scanErr = errors.New("scan failed")
				return r
			}(),
			queryErr: nil,
			want:     nil,
			wantErr:  "scan failed",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			pool := new(mockPgxPool)
			repo := NewPgxMonthlyAverageRepository(pool)

			ctx := context.Background()

			var rows pgx.Rows
			if tc.rows != nil {
				rows = tc.rows
			}

			expectedArgs := buildExpectedArgs(tc.month, tc.year)
			pool.On("Query", ctx, mock.AnythingOfType("string"),
				mock.MatchedBy(argsMatcher(expectedArgs)),
			).Return(rows, tc.queryErr)

			got, err := repo.List(ctx, tc.month, tc.year)

			if tc.wantErr != "" {
				assert.Error(t, err)
				assert.Equal(t, tc.wantErr, err.Error())
				assert.Nil(t, got)
			} else {
				assert.NoError(t, err)
				if tc.want == nil {
					assert.Nil(t, got)
				} else {
					assertMonthlyAverageSliceEqual(t, tc.want, got)
				}
			}

			pool.AssertExpectations(t)
		})
	}
}

// ---------------------------------------------------------------------------
// TestPgxMonthlyAverageRepository_UpsertAverage
// ---------------------------------------------------------------------------

func TestPgxMonthlyAverageRepository_UpsertAverage(t *testing.T) {
	t.Parallel()

	type upsertCase struct {
		name          string
		month         int
		year          int
		averageAmount float64
		expenseCount  int
		execErr       error
		wantErr       string
	}

	cases := []upsertCase{
		{
			name:          "success",
			month:         1,
			year:          2024,
			averageAmount: 150.05,
			expenseCount:  10,
			execErr:       nil,
			wantErr:       "",
		},
		{
			name:          "exec_error",
			month:         1,
			year:          2024,
			averageAmount: 150.05,
			expenseCount:  10,
			execErr:       errors.New("upsert failed"),
			wantErr:       "upsert failed",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			pool := new(mockPgxPool)
			repo := NewPgxMonthlyAverageRepository(pool)

			ctx := context.Background()

			pool.On("Exec", ctx, mock.AnythingOfType("string"),
				[]any{tc.month, tc.year, tc.averageAmount, tc.expenseCount},
			).Return(pgconn.CommandTag{}, tc.execErr)

			err := repo.UpsertAverage(ctx, tc.month, tc.year, tc.averageAmount, tc.expenseCount)

			if tc.wantErr != "" {
				assert.Error(t, err)
				assert.Equal(t, tc.wantErr, err.Error())
			} else {
				assert.NoError(t, err)
			}

			pool.AssertExpectations(t)
		})
	}
}

// ---------------------------------------------------------------------------
// Test helpers
// ---------------------------------------------------------------------------

func intPtr(v int) *int {
	return &v
}

// buildExpectedArgs builds the expected SQL args slice for List queries.
func buildExpectedArgs(month, year *int) []any {
	var args []any
	if month != nil {
		args = append(args, *month)
	}
	if year != nil {
		args = append(args, *year)
	}
	return args
}

// argsMatcher returns a testify MatchedBy function that compares two slices of args.
func argsMatcher(expected []any) func([]any) bool {
	return func(got []any) bool {
		if len(got) != len(expected) {
			return false
		}
		for i := range got {
			if got[i] != expected[i] {
				return false
			}
		}
		return true
	}
}
