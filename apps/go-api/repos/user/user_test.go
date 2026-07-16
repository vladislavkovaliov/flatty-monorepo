package user

import (
	"context"
	"errors"
	"reflect"
	"testing"
	"time"

	userdomain "flatty-budget/go-api/domains/user"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
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
// When scanErr is set Scan returns it immediately without copying values.
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

func strPtr(s string) *string {
	return &s
}

func assertUserEqual(t *testing.T, want, got *userdomain.User) {
	t.Helper()
	assert.Equal(t, want.ID(), got.ID())
	assert.Equal(t, want.Name(), got.Name())
	assert.Equal(t, want.Email(), got.Email())
	assert.Equal(t, want.EmailVerified(), got.EmailVerified())
	if want.Image() == nil {
		assert.Nil(t, got.Image())
	} else {
		require.NotNil(t, got.Image())
		assert.Equal(t, *want.Image(), *got.Image())
	}
	assert.True(t, want.CreatedAt().Equal(got.CreatedAt()), "CreatedAt mismatch")
	assert.True(t, want.UpdatedAt().Equal(got.UpdatedAt()), "UpdatedAt mismatch")
}

func assertUserSliceEqual(t *testing.T, want, got []*userdomain.User) {
	t.Helper()
	require.Equal(t, len(want), len(got))
	for i := range want {
		assertUserEqual(t, want[i], got[i])
	}
}

// ---------------------------------------------------------------------------
// TestPgxRepository_List
// ---------------------------------------------------------------------------

func TestPgxRepository_List(t *testing.T) {
	t.Parallel()

	now := time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)

	type listCase struct {
		name      string
		rows      *mockRows
		queryErr  error
		wantUsers []*userdomain.User
		wantErr   string
	}

	cases := []listCase{
		{
			name: "success",
			rows: newMockRows([][]any{
				{"user-1", "Alice", "alice@test.com", true, strPtr("alice.jpg"), now, now},
				{"user-2", "Bob", "bob@test.com", false, (*string)(nil), now, now},
			}),
			queryErr: nil,
			wantUsers: []*userdomain.User{
				userdomain.NewUser("user-1", "Alice", "alice@test.com", true, strPtr("alice.jpg"), now, now),
				userdomain.NewUser("user-2", "Bob", "bob@test.com", false, nil, now, now),
			},
			wantErr: "",
		},
		{
			name:      "empty",
			rows:      newMockRows(nil),
			queryErr:  nil,
			wantUsers: nil,
			wantErr:   "",
		},
		{
			name:      "query_error",
			rows:      nil,
			queryErr:  errors.New("connection failed"),
			wantUsers: nil,
			wantErr:   "connection failed",
		},
		{
			name: "scan_err_no_rows",
			rows: func() *mockRows {
				r := newMockRows([][]any{
					{"user-1", "Alice", "alice@test.com", true, strPtr("alice.jpg"), now, now},
				})
				r.scanErr = pgx.ErrNoRows
				return r
			}(),
			queryErr:  nil,
			wantUsers: nil,
			wantErr:   "",
		},
		{
			name: "scan_other_error",
			rows: func() *mockRows {
				r := newMockRows([][]any{
					{"user-1", "Alice", "alice@test.com", true, strPtr("alice.jpg"), now, now},
				})
				r.scanErr = errors.New("scan failed")
				return r
			}(),
			queryErr: nil,
			// When Scan returns a non-ErrNoRows error, the code falls through and
			// appends a user with the zero-valued variables (which were declared
			// before Scan and never assigned). This documents the current behaviour.
			wantUsers: []*userdomain.User{
				userdomain.NewUser("", "", "", false, nil, time.Time{}, time.Time{}),
			},
			wantErr: "",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			pool := new(mockPgxPool)
			repo := NewPgxRepository(pool)

			ctx := context.Background()
			limit := 10
			offset := 0

			// Set up Query expectation (called for all cases except query_error
			// where we still call it but it returns err).
			var rows pgx.Rows
			if tc.rows != nil {
				rows = tc.rows
			}
			pool.On("Query", ctx, mock.AnythingOfType("string"), []any{limit, offset}).
				Return(rows, tc.queryErr)

			users, err := repo.List(ctx, limit, offset)

			if tc.wantErr != "" {
				assert.Error(t, err)
				assert.Equal(t, tc.wantErr, err.Error())
				assert.Nil(t, users)
			} else {
				assert.NoError(t, err)
				if tc.wantUsers == nil {
					assert.Nil(t, users)
				} else {
					assertUserSliceEqual(t, tc.wantUsers, users)
				}
			}

			pool.AssertExpectations(t)
		})
	}
}

// ---------------------------------------------------------------------------
// TestPgxRepository_GetUserByID
// ---------------------------------------------------------------------------

func TestPgxRepository_GetUserByID(t *testing.T) {
	t.Parallel()

	now := time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)

	type getCase struct {
		name     string
		row      *mockRow
		wantUser *userdomain.User
		wantErr  string
	}

	cases := []getCase{
		{
			name: "success",
			row: newMockRow([]any{
				"user-1", "Alice", "alice@test.com", true, strPtr("alice.jpg"), now, now,
			}),
			wantUser: userdomain.NewUser("user-1", "Alice", "alice@test.com", true, strPtr("alice.jpg"), now, now),
			wantErr:  "",
		},
		{
			name:     "not_found",
			row:      newMockRowWithError(pgx.ErrNoRows),
			wantUser: nil,
			wantErr:  "",
		},
		{
			name:     "query_error",
			row:      newMockRowWithError(errors.New("db error")),
			wantUser: nil,
			wantErr:  "db error",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			pool := new(mockPgxPool)
			repo := NewPgxRepository(pool)

			ctx := context.Background()
			userID := "some-user-id"

			pool.On("QueryRow", ctx, mock.AnythingOfType("string"), []any{userID}).
				Return(tc.row)

			user, err := repo.GetUserByID(ctx, userID)

			if tc.wantErr != "" {
				assert.Error(t, err)
				assert.Equal(t, tc.wantErr, err.Error())
				assert.Nil(t, user)
			} else {
				assert.NoError(t, err)
				if tc.wantUser == nil {
					assert.Nil(t, user)
				} else {
					assertUserEqual(t, tc.wantUser, user)
				}
			}

			pool.AssertExpectations(t)
		})
	}
}
