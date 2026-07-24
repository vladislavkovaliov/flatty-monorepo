package user_settings

import (
	"context"
	"errors"
	"reflect"
	"testing"
	"time"

	user_settings_domain "flatty-budget/go-api/domains/user_settings"

	"github.com/jackc/pgx/v5"
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
		row      *mockRow
		want     *user_settings_domain.UserSettings
		wantErr  string
	}

	cases := []getCase{
		{
			name: "success",
			row: newMockRow([]any{
				"user-1", "en", "dark", "America/New_York", "MM/DD/YYYY", now, now,
			}),
			want: user_settings_domain.NewUserSettings("user-1", "en", "dark", "America/New_York", "MM/DD/YYYY", now, now),
			wantErr: "",
		},
		{
			name:    "not_found",
			row:     newMockRowWithError(pgx.ErrNoRows),
			want:    nil,
			wantErr: "",
		},
		{
			name:    "query_error",
			row:     newMockRowWithError(errors.New("db error")),
			want:    nil,
			wantErr: "db error",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			pool := new(mockPgxPool)
			repo := NewPgxRepository(pool)

			ctx := context.Background()
			userID := "user-1"

			pool.On("QueryRow", ctx, mock.AnythingOfType("string"), []any{userID}).
				Return(tc.row)

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
		name    string
		userID  string
		input   *user_settings_domain.UserSettingsInput
		row     *mockRow
		want    *user_settings_domain.UserSettings
		wantErr string
	}

	cases := []upsertCase{
		{
			name:   "success",
			userID: "user-1",
			input:  user_settings_domain.NewUserSettingsInput("en", "dark", "America/New_York", "MM/DD/YYYY"),
			row: newMockRow([]any{
				"user-1", "en", "dark", "America/New_York", "MM/DD/YYYY", now, now,
			}),
			want: user_settings_domain.NewUserSettings("user-1", "en", "dark", "America/New_York", "MM/DD/YYYY", now, now),
			wantErr: "",
		},
		{
			name:    "query_error",
			userID:  "user-1",
			input:   user_settings_domain.NewUserSettingsInput("en", "dark", "America/New_York", "MM/DD/YYYY"),
			row:     newMockRowWithError(errors.New("db error")),
			want:    nil,
			wantErr: "db error",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			pool := new(mockPgxPool)
			repo := NewPgxRepository(pool)

			ctx := context.Background()

			pool.On("QueryRow", ctx, mock.AnythingOfType("string"), []any{
				tc.userID,
				tc.input.Language(),
				tc.input.Theme(),
				tc.input.Timezone(),
				tc.input.DateFormat(),
			}).Return(tc.row)

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