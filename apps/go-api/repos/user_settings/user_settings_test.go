package user_settings

import (
	"context"
	"errors"
	"reflect"
	"testing"
	"time"

	user_settings_domain "flatty-budget/go-api/domains/user_settings"

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

func assertUserSettingsEqual(t *testing.T, want, got *user_settings_domain.UserSettings) {
	t.Helper()
	assert.Equal(t, want.UserID(), got.UserID())
	assert.Equal(t, want.Language(), got.Language())
	assert.Equal(t, want.Theme(), got.Theme())
	assert.Equal(t, want.Timezone(), got.Timezone())
	assert.Equal(t, want.DateFormat(), got.DateFormat())
	assert.True(t, want.CreatedAt().Equal(got.CreatedAt()), "CreatedAt mismatch")
	assert.True(t, want.UpdatedAt().Equal(got.UpdatedAt()), "UpdatedAt mismatch")
}

func TestPgxRepository_GetByUserID(t *testing.T) {
	t.Parallel()

	now := time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)

	type getCase struct {
		name     string
		rows     *mockRows
		queryErr error
		want     *user_settings_domain.UserSettings
		wantErr  string
	}

	cases := []getCase{
		{
			name: "success",
			rows: newMockRows([][]any{
				{"user-1", "en", "dark", "America/New_York", "MM/DD/YYYY", now, now},
			}),
			want:    user_settings_domain.NewUserSettings("user-1", "en", "dark", "America/New_York", "MM/DD/YYYY", now, now),
			wantErr: "",
		},
		{
			name:    "not_found",
			rows:    newMockRows(nil),
			want:    nil,
			wantErr: "",
		},
		{
			name:     "query_error",
			queryErr: errors.New("db error"),
			want:     nil,
			wantErr:  "db error",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			pool := new(mockPgxPool)
			repo := NewPgxRepository(pool)

			ctx := context.Background()
			userID := "user-1"

			var rows pgx.Rows
			if tc.rows != nil {
				rows = tc.rows
			}
			pool.On("Query", ctx, mock.AnythingOfType("string"), []any{userID}).
				Return(rows, tc.queryErr)

			got, err := repo.GetByUserID(ctx, userID)

			if tc.wantErr != "" {
				assert.Error(t, err)
				assert.Equal(t, tc.wantErr, err.Error())
				assert.Nil(t, got)
			} else {
				assert.NoError(t, err)
				if tc.want == nil {
					assert.Nil(t, got)
				} else {
					assertUserSettingsEqual(t, tc.want, got)
				}
			}

			pool.AssertExpectations(t)
		})
	}
}

func TestPgxRepository_Upsert(t *testing.T) {
	t.Parallel()

	now := time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)

	type upsertCase struct {
		name     string
		userID   string
		input    *user_settings_domain.UserSettingsInput
		rows     *mockRows
		queryErr error
		want     *user_settings_domain.UserSettings
		wantErr  string
	}

	cases := []upsertCase{
		{
			name:   "success",
			userID: "user-1",
			input:  user_settings_domain.NewUserSettingsInput("en", "dark", "America/New_York", "MM/DD/YYYY"),
			rows: newMockRows([][]any{
				{"user-1", "en", "dark", "America/New_York", "MM/DD/YYYY", now, now},
			}),
			want:    user_settings_domain.NewUserSettings("user-1", "en", "dark", "America/New_York", "MM/DD/YYYY", now, now),
			wantErr: "",
		},
		{
			name:     "query_error",
			userID:   "user-1",
			input:    user_settings_domain.NewUserSettingsInput("en", "dark", "America/New_York", "MM/DD/YYYY"),
			queryErr: errors.New("db error"),
			want:     nil,
			wantErr:  "db error",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			pool := new(mockPgxPool)
			repo := NewPgxRepository(pool)

			ctx := context.Background()

			var rows pgx.Rows
			if tc.rows != nil {
				rows = tc.rows
			}
			pool.On("Query", ctx, mock.AnythingOfType("string"), []any{
				tc.userID,
				tc.input.Language(),
				tc.input.Theme(),
				tc.input.Timezone(),
				tc.input.DateFormat(),
			}).Return(rows, tc.queryErr)

			got, err := repo.Upsert(ctx, tc.userID, tc.input)

			if tc.wantErr != "" {
				assert.Error(t, err)
				assert.Equal(t, tc.wantErr, err.Error())
				assert.Nil(t, got)
			} else {
				assert.NoError(t, err)
				assertUserSettingsEqual(t, tc.want, got)
			}

			pool.AssertExpectations(t)
		})
	}
}
