package expenses

import (
	"context"
	"errors"
	"reflect"
	"testing"
	"time"

	expensedomain "flatty-budget/go-api/domains/expenses"
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

// mockRow implements pgx.Row for testing.
type mockRow struct {
	scanValues []any
	scanErr    error
}

func newMockRow(values []any) *mockRow {
	return &mockRow{scanValues: values}
}

func newMockRowWithError(err error) *mockRow {
	return &mockRow{scanErr: err}
}

func (m *mockRow) Scan(dest ...any) error {
	if m.scanErr != nil {
		return m.scanErr
	}
	for i, d := range dest {
		if i >= len(m.scanValues) {
			break
		}
		v := reflect.ValueOf(d)
		if v.Kind() != reflect.Ptr {
			continue
		}
		srcVal := reflect.ValueOf(m.scanValues[i])
		if srcVal.IsValid() {
			v.Elem().Set(srcVal)
		}
	}
	return nil
}

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

func assertExpenseEqual(t *testing.T, want, got *expensedomain.Expense) {
	t.Helper()
	assert.Equal(t, want.ID(), got.ID())
	assert.Equal(t, want.ResidentLocationID(), got.ResidentLocationID())
	assert.Equal(t, want.CategoryID(), got.CategoryID())
	assert.Equal(t, want.Amount(), got.Amount())
	assert.Equal(t, want.Month(), got.Month())
	assert.Equal(t, want.Year(), got.Year())
	assert.True(t, want.CreatedAt().Equal(got.CreatedAt()), "CreatedAt mismatch")
	assert.True(t, want.UpdatedAt().Equal(got.UpdatedAt()), "UpdatedAt mismatch")
}

func assertExpenseSliceEqual(t *testing.T, want, got []*expensedomain.Expense) {
	t.Helper()
	assert.Equal(t, len(want), len(got))
	for i := range want {
		assertExpenseEqual(t, want[i], got[i])
	}
}

// ---------------------------------------------------------------------------
// TestPgxRepository_Count
// ---------------------------------------------------------------------------

func TestPgxRepository_Count(t *testing.T) {
	t.Parallel()

	type countCase struct {
		name    string
		row     *mockRow
		want    int
		wantErr string
	}

	cases := []countCase{
		{
			name: "success",
			row:  newMockRow([]any{10}),
			want: 10,
		},
		{
			name:    "query_error",
			row:     newMockRowWithError(errors.New("db error")),
			want:    0,
			wantErr: "db error",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			pool := new(mockPgxPool)
			repo := NewPgxRepository(pool)

			ctx := context.Background()

			pool.On("QueryRow", ctx, mock.AnythingOfType("string"), ([]any)(nil)).
				Return(tc.row)

			got, err := repo.Count(ctx)

			if tc.wantErr != "" {
				assert.Error(t, err)
				assert.Equal(t, tc.wantErr, err.Error())
				assert.Equal(t, tc.want, got)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.want, got)
			}

			pool.AssertExpectations(t)
		})
	}
}

// ---------------------------------------------------------------------------
// TestPgxRepository_List
// ---------------------------------------------------------------------------

func TestPgxRepository_List(t *testing.T) {
	t.Parallel()

	now := time.Date(2024, 6, 15, 10, 30, 0, 0, time.UTC)

	type listCase struct {
		name         string
		rows         *mockRows
		queryErr     error
		wantExpenses []*expensedomain.Expense
		wantErr      string
	}

	cases := []listCase{
		{
			name: "success",
			rows: newMockRows([][]any{
				{int64(1), int64(10), int64(20), 150.50, 6, 2024, now, now},
				{int64(2), int64(10), int64(21), 75.00, 6, 2024, now, now},
			}),
			queryErr: nil,
			wantExpenses: []*expensedomain.Expense{
				expensedomain.NewExpense(1, 10, 20, 150.50, 6, 2024, now, now),
				expensedomain.NewExpense(2, 10, 21, 75.00, 6, 2024, now, now),
			},
			wantErr: "",
		},
		{
			name:         "empty",
			rows:         newMockRows(nil),
			queryErr:     nil,
			wantExpenses: nil,
			wantErr:      "",
		},
		{
			name:         "query_error",
			rows:         nil,
			queryErr:     errors.New("connection failed"),
			wantExpenses: nil,
			wantErr:      "connection failed",
		},
		{
			name: "scan_error",
			rows: func() *mockRows {
				r := newMockRows([][]any{
					{int64(1), int64(10), int64(20), 150.50, 6, 2024, now, now},
				})
				r.scanErr = errors.New("scan failed")
				return r
			}(),
			queryErr:     nil,
			wantExpenses: nil,
			wantErr:      "scan failed",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			pool := new(mockPgxPool)
			repo := NewPgxRepository(pool)

			ctx := context.Background()
			limit := 10
			offset := 0

			var rows pgx.Rows
			if tc.rows != nil {
				rows = tc.rows
			}
			pool.On("Query", ctx, mock.AnythingOfType("string"), []any{limit, offset}).
				Return(rows, tc.queryErr)

			expenses, err := repo.List(ctx, limit, offset)

			if tc.wantErr != "" {
				assert.Error(t, err)
				assert.Equal(t, tc.wantErr, err.Error())
				assert.Nil(t, expenses)
			} else {
				assert.NoError(t, err)
				if tc.wantExpenses == nil {
					assert.Nil(t, expenses)
				} else {
					assertExpenseSliceEqual(t, tc.wantExpenses, expenses)
				}
			}

			pool.AssertExpectations(t)
		})
	}
}

// ---------------------------------------------------------------------------
// TestPgxRepository_GetByID
// ---------------------------------------------------------------------------

func TestPgxRepository_GetByID(t *testing.T) {
	t.Parallel()

	now := time.Date(2024, 6, 15, 10, 30, 0, 0, time.UTC)

	type getCase struct {
		name        string
		row         *mockRow
		id          int64
		wantExpense *expensedomain.Expense
		wantErr     string
	}

	cases := []getCase{
		{
			name: "success",
			row:  newMockRow([]any{int64(1), int64(10), int64(20), 150.50, 6, 2024, now, now}),
			id:   1,
			wantExpense: expensedomain.NewExpense(1, 10, 20, 150.50, 6, 2024, now, now),
			wantErr: "",
		},
		{
			name:        "not_found",
			row:         newMockRowWithError(pgx.ErrNoRows),
			id:          999,
			wantExpense: nil,
			wantErr:     "expense with id 999 not found: no rows in result set",
		},
		{
			name:        "query_error",
			row:         newMockRowWithError(errors.New("db error")),
			id:          1,
			wantExpense: nil,
			wantErr:     "db error",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			pool := new(mockPgxPool)
			repo := NewPgxRepository(pool)

			ctx := context.Background()

			pool.On("QueryRow", ctx, mock.AnythingOfType("string"), []any{tc.id}).
				Return(tc.row)

			expense, err := repo.GetByID(ctx, tc.id)

			if tc.wantErr != "" {
				assert.Error(t, err)
				assert.Equal(t, tc.wantErr, err.Error())
				assert.Nil(t, expense)
			} else {
				assert.NoError(t, err)
				assertExpenseEqual(t, tc.wantExpense, expense)
			}

			pool.AssertExpectations(t)
		})
	}
}

// ---------------------------------------------------------------------------
// TestPgxRepository_Create
// ---------------------------------------------------------------------------

func TestPgxRepository_Create(t *testing.T) {
	t.Parallel()

	now := time.Date(2024, 6, 15, 10, 30, 0, 0, time.UTC)

	type createCase struct {
		name         string
		row          *mockRow
		input        *expensedomain.ExpenseInput
		wantExpense  *expensedomain.Expense
		wantErr      string
	}

	cases := []createCase{
		{
			name: "success",
			row:  newMockRow([]any{int64(1), int64(10), int64(20), 150.50, 6, 2024, now, now}),
			input: expensedomain.NewExpenseInput(10, 20, 150.50, 6, 2024),
			wantExpense: expensedomain.NewExpense(1, 10, 20, 150.50, 6, 2024, now, now),
			wantErr: "",
		},
		{
			name:    "query_error",
			row:     newMockRowWithError(errors.New("insert failed")),
			input:   expensedomain.NewExpenseInput(10, 20, 150.50, 6, 2024),
			wantExpense: nil,
			wantErr: "insert failed",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			pool := new(mockPgxPool)
			repo := NewPgxRepository(pool)

			ctx := context.Background()

			pool.On("QueryRow", ctx, mock.AnythingOfType("string"),
				[]any{tc.input.ResidentLocationID(), tc.input.CategoryID(), tc.input.Amount(), tc.input.Month(), tc.input.Year()},
			).Return(tc.row)

			expense, err := repo.Create(ctx, tc.input)

			if tc.wantErr != "" {
				assert.Error(t, err)
				assert.Equal(t, tc.wantErr, err.Error())
				assert.Nil(t, expense)
			} else {
				assert.NoError(t, err)
				assertExpenseEqual(t, tc.wantExpense, expense)
			}

			pool.AssertExpectations(t)
		})
	}
}

// ---------------------------------------------------------------------------
// TestPgxRepository_Update
// ---------------------------------------------------------------------------

func TestPgxRepository_Update(t *testing.T) {
	t.Parallel()

	now := time.Date(2024, 6, 15, 10, 30, 0, 0, time.UTC)

	type updateCase struct {
		name        string
		row         *mockRow
		id          int64
		input       *expensedomain.ExpenseInput
		wantExpense *expensedomain.Expense
		wantErr     string
	}

	cases := []updateCase{
		{
			name: "success",
			row:  newMockRow([]any{int64(1), int64(10), int64(20), 200.00, 7, 2024, now, now}),
			id:   1,
			input: expensedomain.NewExpenseInput(10, 20, 200.00, 7, 2024),
			wantExpense: expensedomain.NewExpense(1, 10, 20, 200.00, 7, 2024, now, now),
			wantErr: "",
		},
		{
			name:    "not_found",
			row:     newMockRowWithError(pgx.ErrNoRows),
			id:      999,
			input:   expensedomain.NewExpenseInput(10, 20, 200.00, 7, 2024),
			wantExpense: nil,
			wantErr: "expense with id 999 not found: no rows in result set",
		},
		{
			name:    "query_error",
			row:     newMockRowWithError(errors.New("update failed")),
			id:      1,
			input:   expensedomain.NewExpenseInput(10, 20, 200.00, 7, 2024),
			wantExpense: nil,
			wantErr: "update failed",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			pool := new(mockPgxPool)
			repo := NewPgxRepository(pool)

			ctx := context.Background()

			pool.On("QueryRow", ctx, mock.AnythingOfType("string"),
				[]any{tc.input.ResidentLocationID(), tc.input.CategoryID(), tc.input.Amount(), tc.input.Month(), tc.input.Year(), tc.id},
			).Return(tc.row)

			expense, err := repo.Update(ctx, tc.id, tc.input)

			if tc.wantErr != "" {
				assert.Error(t, err)
				assert.Equal(t, tc.wantErr, err.Error())
				assert.Nil(t, expense)
			} else {
				assert.NoError(t, err)
				assertExpenseEqual(t, tc.wantExpense, expense)
			}

			pool.AssertExpectations(t)
		})
	}
}

// ---------------------------------------------------------------------------
// TestPgxRepository_Delete
// ---------------------------------------------------------------------------

func TestPgxRepository_Delete(t *testing.T) {
	t.Parallel()

	type deleteCase struct {
		name    string
		row     *mockRow
		id      int64
		want    int64
		wantErr string
	}

	cases := []deleteCase{
		{
			name: "success",
			row:  newMockRow([]any{int64(1)}),
			id:   1,
			want: 1,
		},
		{
			name:    "not_found",
			row:     newMockRowWithError(pgx.ErrNoRows),
			id:      999,
			want:    -1,
			wantErr: "expense with id 999 not found: no rows in result set",
		},
		{
			name:    "query_error",
			row:     newMockRowWithError(errors.New("delete failed")),
			id:      1,
			want:    -1,
			wantErr: "delete failed",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			pool := new(mockPgxPool)
			repo := NewPgxRepository(pool)

			ctx := context.Background()

			pool.On("QueryRow", ctx, mock.AnythingOfType("string"), []any{tc.id}).
				Return(tc.row)

			got, err := repo.Delete(ctx, tc.id)

			if tc.wantErr != "" {
				assert.Error(t, err)
				assert.Equal(t, tc.wantErr, err.Error())
				assert.Equal(t, tc.want, got)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.want, got)
			}

			pool.AssertExpectations(t)
		})
	}
}
