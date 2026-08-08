package applications

import (
	"context"
	"errors"
	"reflect"
	"testing"
	"time"

	applicationsdomain "flatty-budget/go-api/domains/applications"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

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

type mockRow struct {
	scanValues []any
	scanErr    error
}

func newMockRow(values []any) *mockRow       { return &mockRow{scanValues: values} }
func newMockRowWithError(err error) *mockRow { return &mockRow{scanErr: err} }

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

type mockRows struct {
	rows [][]any
	idx  int
	err  error
}

func (m *mockRows) Next() bool                                   { return m.idx < len(m.rows) }
func (m *mockRows) Close()                                       {}
func (m *mockRows) Err() error                                   { return m.err }
func (m *mockRows) CommandTag() pgconn.CommandTag                { return pgconn.CommandTag{} }
func (m *mockRows) FieldDescriptions() []pgconn.FieldDescription { return nil }
func (m *mockRows) Values() ([]any, error)                       { return nil, nil }
func (m *mockRows) RawValues() [][]byte                          { return nil }
func (m *mockRows) Conn() *pgx.Conn                              { return nil }

func (m *mockRows) Scan(dest ...any) error {
	if m.err != nil {
		return m.err
	}
	if m.idx >= len(m.rows) {
		return pgx.ErrNoRows
	}
	vals := m.rows[m.idx]
	m.idx++
	for i, d := range dest {
		if i >= len(vals) {
			break
		}
		v := reflect.ValueOf(d)
		if v.Kind() != reflect.Ptr {
			continue
		}
		srcVal := reflect.ValueOf(vals[i])
		if srcVal.IsValid() {
			v.Elem().Set(srcVal)
		}
	}
	return nil
}

func appRow(now time.Time) []any {
	return []any{int64(1), "resident", "development", "resident", "styles",
		"http://localhost:8082", "/external-resident", "/", now, now}
}

func require_Len(t *testing.T, apps []*applicationsdomain.Application, want int) {
	t.Helper()
	assert.Equal(t, want, len(apps))
}

func TestPgxRepository_ListByEnv(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, 7, 31, 10, 0, 0, 0, time.UTC)

	t.Run("success", func(t *testing.T) {
		pool := new(mockPgxPool)
		rows := &mockRows{rows: [][]any{appRow(now)}}
		ctx := context.Background()

		pool.On("Query", ctx, mock.AnythingOfType("string"), []any{"development"}).Return(rows, nil)

		repo := NewPgxRepository(pool)
		apps, err := repo.ListByEnv(ctx, "development")

		assert.NoError(t, err)
		require_Len(t, apps, 1)
		assert.Equal(t, "resident", apps[0].Name())
		assert.Equal(t, "development", apps[0].Env())
		pool.AssertExpectations(t)
	})

	t.Run("query_error", func(t *testing.T) {
		pool := new(mockPgxPool)
		ctx := context.Background()

		pool.On("Query", ctx, mock.AnythingOfType("string"), []any{"development"}).Return(nil, errors.New("db error"))

		repo := NewPgxRepository(pool)
		_, err := repo.ListByEnv(ctx, "development")

		assert.EqualError(t, err, "db error")
		pool.AssertExpectations(t)
	})
}

func TestPgxRepository_ListAll(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, 7, 31, 10, 0, 0, 0, time.UTC)
	pool := new(mockPgxPool)
	rows := &mockRows{rows: [][]any{appRow(now)}}
	ctx := context.Background()

	pool.On("Query", ctx, mock.AnythingOfType("string"), mock.Anything).Return(rows, nil)

	repo := NewPgxRepository(pool)
	apps, err := repo.ListAll(ctx)

	assert.NoError(t, err)
	require_Len(t, apps, 1)
	assert.Equal(t, "resident", apps[0].Name())
	pool.AssertExpectations(t)
}

func TestPgxRepository_Create(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, 7, 31, 10, 0, 0, 0, time.UTC)
	pool := new(mockPgxPool)
	ctx := context.Background()

	input := applicationsdomain.NewApplicationInput("app", "development", "app", "styles",
		"http://localhost:8080", "/external-app", "/")

	pool.On("QueryRow", ctx, mock.AnythingOfType("string"), []any{
		"app", "development", "app", "styles", "http://localhost:8080", "/external-app", "/",
	}).Return(newMockRow([]any{int64(1), "app", "development", "app", "styles",
		"http://localhost:8080", "/external-app", "/", now, now}))

	repo := NewPgxRepository(pool)
	app, err := repo.Create(ctx, input)

	assert.NoError(t, err)
	assert.Equal(t, int64(1), app.ID())
	assert.Equal(t, "app", app.Name())
	pool.AssertExpectations(t)
}

func TestPgxRepository_Update(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, 7, 31, 10, 0, 0, 0, time.UTC)

	t.Run("success", func(t *testing.T) {
		pool := new(mockPgxPool)
		ctx := context.Background()
		input := applicationsdomain.NewApplicationInput("resident", "development", "resident", "styles",
			"http://localhost:8082", "/external-resident", "/")

		pool.On("QueryRow", ctx, mock.AnythingOfType("string"), []any{
			"resident", "development", "resident", "styles", "http://localhost:8082", "/external-resident", "/", int64(3),
		}).Return(newMockRow(appRow(now)))

		repo := NewPgxRepository(pool)
		app, err := repo.Update(ctx, 3, input)

		assert.NoError(t, err)
		assert.Equal(t, int64(1), app.ID())
		pool.AssertExpectations(t)
	})

	t.Run("not_found", func(t *testing.T) {
		pool := new(mockPgxPool)
		ctx := context.Background()
		input := applicationsdomain.NewApplicationInput("resident", "development", "resident", "styles",
			"http://localhost:8082", "/external-resident", "/")

		pool.On("QueryRow", ctx, mock.AnythingOfType("string"), mock.Anything).
			Return(newMockRowWithError(pgx.ErrNoRows))

		repo := NewPgxRepository(pool)
		_, err := repo.Update(ctx, 999, input)

		assert.ErrorIs(t, err, pgx.ErrNoRows)
		assert.Contains(t, err.Error(), "application with id 999 not found")
		pool.AssertExpectations(t)
	})
}

func TestPgxRepository_Delete(t *testing.T) {
	t.Parallel()

	t.Run("success", func(t *testing.T) {
		pool := new(mockPgxPool)
		ctx := context.Background()

		pool.On("QueryRow", ctx, mock.AnythingOfType("string"), []any{int64(2)}).
			Return(newMockRow([]any{int64(2)}))

		repo := NewPgxRepository(pool)
		id, err := repo.Delete(ctx, 2)

		assert.NoError(t, err)
		assert.Equal(t, int64(2), id)
		pool.AssertExpectations(t)
	})

	t.Run("not_found", func(t *testing.T) {
		pool := new(mockPgxPool)
		ctx := context.Background()

		pool.On("QueryRow", ctx, mock.AnythingOfType("string"), []any{int64(999)}).
			Return(newMockRowWithError(pgx.ErrNoRows))

		repo := NewPgxRepository(pool)
		_, err := repo.Delete(ctx, 999)

		assert.ErrorIs(t, err, pgx.ErrNoRows)
		assert.Contains(t, err.Error(), "application with id 999 not found")
		pool.AssertExpectations(t)
	})
}
