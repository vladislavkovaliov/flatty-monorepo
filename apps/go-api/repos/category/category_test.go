package category

import (
	"context"
	"errors"
	"reflect"
	"testing"
	"time"

	categorydomain "flatty-budget/go-api/domains/category"
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

func assertCategoryEqual(t *testing.T, want, got *categorydomain.Category) {
	t.Helper()
	assert.Equal(t, want.ID(), got.ID())
	assert.Equal(t, want.Name(), got.Name())
	assert.Equal(t, want.Description(), got.Description())
	assert.True(t, want.CreatedAt().Equal(got.CreatedAt()), "CreatedAt mismatch")
	assert.True(t, want.UpdatedAt().Equal(got.UpdatedAt()), "UpdatedAt mismatch")
}

func assertCategorySliceEqual(t *testing.T, want, got []*categorydomain.Category) {
	t.Helper()
	assert.Equal(t, len(want), len(got))
	for i := range want {
		assertCategoryEqual(t, want[i], got[i])
	}
}

// ---------------------------------------------------------------------------
// TestPgxRepository_Count
// ---------------------------------------------------------------------------

func TestPgxRepository_Count(t *testing.T) {
	t.Parallel()

	type countCase struct {
		name     string
		row      *mockRow
		want     int
		wantErr  string
	}

	cases := []countCase{
		{
			name: "success",
			row:  newMockRow([]any{42}),
			want: 42,
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
		name          string
		rows          *mockRows
		queryErr      error
		wantCategories []*categorydomain.Category
		wantErr       string
	}

	cases := []listCase{
		{
			name: "success",
			rows: newMockRows([][]any{
				{int64(1), "Utilities", "Gas & electricity", now, now},
				{int64(2), "Rent", "Monthly rent", now, now},
			}),
			queryErr: nil,
			wantCategories: []*categorydomain.Category{
				categorydomain.NewCategory(1, "Utilities", "Gas & electricity", now, now),
				categorydomain.NewCategory(2, "Rent", "Monthly rent", now, now),
			},
			wantErr: "",
		},
		{
			name:          "empty",
			rows:          newMockRows(nil),
			queryErr:      nil,
			wantCategories: nil,
			wantErr:       "",
		},
		{
			name:          "query_error",
			rows:          nil,
			queryErr:      errors.New("connection failed"),
			wantCategories: nil,
			wantErr:       "connection failed",
		},
		{
			name: "scan_error",
			rows: func() *mockRows {
				r := newMockRows([][]any{
					{int64(1), "Utilities", "Gas & electricity", now, now},
				})
				r.scanErr = errors.New("scan failed")
				return r
			}(),
			queryErr:      nil,
			wantCategories: nil,
			wantErr:       "scan failed",
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

			categories, err := repo.List(ctx, limit, offset)

			if tc.wantErr != "" {
				assert.Error(t, err)
				assert.Equal(t, tc.wantErr, err.Error())
				assert.Nil(t, categories)
			} else {
				assert.NoError(t, err)
				if tc.wantCategories == nil {
					assert.Nil(t, categories)
				} else {
					assertCategorySliceEqual(t, tc.wantCategories, categories)
				}
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
		input        *categorydomain.CategoryInput
		wantCategory *categorydomain.Category
		wantErr      string
	}

	cases := []createCase{
		{
			name:  "success",
			row:   newMockRow([]any{int64(1), "Utilities", "Gas & electricity", now, now}),
			input: categorydomain.NewCategoryInput("Utilities", "Gas & electricity"),
			wantCategory: categorydomain.NewCategory(1, "Utilities", "Gas & electricity", now, now),
			wantErr: "",
		},
		{
			name:    "query_error",
			row:     newMockRowWithError(errors.New("insert failed")),
			input:   categorydomain.NewCategoryInput("Utilities", "Gas & electricity"),
			wantCategory: nil,
			wantErr: "insert failed",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			pool := new(mockPgxPool)
			repo := NewPgxRepository(pool)

			ctx := context.Background()

			pool.On("QueryRow", ctx, mock.AnythingOfType("string"), []any{tc.input.Name(), tc.input.Description()}).
				Return(tc.row)

			category, err := repo.Create(ctx, tc.input)

			if tc.wantErr != "" {
				assert.Error(t, err)
				assert.Equal(t, tc.wantErr, err.Error())
				assert.Nil(t, category)
			} else {
				assert.NoError(t, err)
				assertCategoryEqual(t, tc.wantCategory, category)
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
		name         string
		row          *mockRow
		id           int64
		input        *categorydomain.CategoryInput
		wantCategory *categorydomain.Category
		wantErr      string
	}

	cases := []updateCase{
		{
			name:  "success",
			row:   newMockRow([]any{int64(1), "Updated Name", "Updated desc", now, now}),
			id:    1,
			input: categorydomain.NewCategoryInput("Updated Name", "Updated desc"),
			wantCategory: categorydomain.NewCategory(1, "Updated Name", "Updated desc", now, now),
			wantErr: "",
		},
		{
			name:    "not_found",
			row:     newMockRowWithError(pgx.ErrNoRows),
			id:      999,
			input:   categorydomain.NewCategoryInput("Updated Name", "Updated desc"),
			wantCategory: nil,
			wantErr: "category with id 999 not found: no rows in result set",
		},
		{
			name:    "query_error",
			row:     newMockRowWithError(errors.New("update failed")),
			id:      1,
			input:   categorydomain.NewCategoryInput("Updated Name", "Updated desc"),
			wantCategory: nil,
			wantErr: "update failed",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			pool := new(mockPgxPool)
			repo := NewPgxRepository(pool)

			ctx := context.Background()

			pool.On("QueryRow", ctx, mock.AnythingOfType("string"), []any{tc.input.Name(), tc.input.Description(), tc.id}).
				Return(tc.row)

			category, err := repo.Update(ctx, tc.id, tc.input)

			if tc.wantErr != "" {
				assert.Error(t, err)
				assert.Equal(t, tc.wantErr, err.Error())
				assert.Nil(t, category)
			} else {
				assert.NoError(t, err)
				assertCategoryEqual(t, tc.wantCategory, category)
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
			wantErr: "category with id 999 not found: no rows in result set",
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
