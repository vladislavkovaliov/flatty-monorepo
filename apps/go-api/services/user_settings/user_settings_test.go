package user_settings

import (
	"context"
	"errors"
	"testing"
	"time"

	user_settings_domain "flatty-budget/go-api/domains/user_settings"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type mockRepo struct {
	mock.Mock
}

func (m *mockRepo) GetByUserID(ctx context.Context, userID string) (*user_settings_domain.UserSettings, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*user_settings_domain.UserSettings), args.Error(1)
}

func (m *mockRepo) Upsert(ctx context.Context, userID string, input *user_settings_domain.UserSettingsInput) (*user_settings_domain.UserSettings, error) {
	args := m.Called(ctx, userID, input)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*user_settings_domain.UserSettings), args.Error(1)
}

func TestService_Get_ReturnsDefaultsWhenNoRow(t *testing.T) {
	t.Parallel()

	repo := new(mockRepo)
	svc := NewService(repo)

	repo.On("GetByUserID", mock.Anything, "user-1").Return(nil, nil)

	got, err := svc.Get(context.Background(), "user-1")

	assert.NoError(t, err)
	assert.NotNil(t, got)
	assert.Equal(t, "user-1", got.UserID())
	assert.Equal(t, "en", got.Language())
	assert.Equal(t, "system", got.Theme())
	assert.Equal(t, "UTC+0", got.Timezone())
	assert.Equal(t, "YYYY-MM-DD", got.DateFormat())
	assert.False(t, got.CreatedAt().IsZero())
	assert.False(t, got.UpdatedAt().IsZero())

	repo.AssertExpectations(t)
}

func TestService_Get_ReturnsExistingSettings(t *testing.T) {
	t.Parallel()

	now := time.Now()
	expected := user_settings_domain.NewUserSettings("user-1", "fr", "dark", "Europe/Paris", "DD/MM/YYYY", now, now)

	repo := new(mockRepo)
	svc := NewService(repo)

	repo.On("GetByUserID", mock.Anything, "user-1").Return(expected, nil)

	got, err := svc.Get(context.Background(), "user-1")

	assert.NoError(t, err)
	assert.Equal(t, expected, got)

	repo.AssertExpectations(t)
}

func TestService_Get_RepoError(t *testing.T) {
	t.Parallel()

	repo := new(mockRepo)
	svc := NewService(repo)

	repo.On("GetByUserID", mock.Anything, "user-1").Return(nil, errors.New("db error"))

	got, err := svc.Get(context.Background(), "user-1")

	assert.Error(t, err)
	assert.Equal(t, "db error", err.Error())
	assert.Nil(t, got)

	repo.AssertExpectations(t)
}

func TestService_Update_Success(t *testing.T) {
	t.Parallel()

	now := time.Now()
	input := user_settings_domain.NewUserSettingsInput("en", "dark", "America/New_York", "MM/DD/YYYY")
	expected := user_settings_domain.NewUserSettings("user-1", "en", "dark", "America/New_York", "MM/DD/YYYY", now, now)

	repo := new(mockRepo)
	svc := NewService(repo)

	repo.On("Upsert", mock.Anything, "user-1", input).Return(expected, nil)

	got, err := svc.Update(context.Background(), "user-1", input)

	assert.NoError(t, err)
	assert.Equal(t, expected, got)

	repo.AssertExpectations(t)
}

func TestService_Update_RepoError(t *testing.T) {
	t.Parallel()

	input := user_settings_domain.NewUserSettingsInput("en", "dark", "America/New_York", "MM/DD/YYYY")

	repo := new(mockRepo)
	svc := NewService(repo)

	repo.On("Upsert", mock.Anything, "user-1", input).Return(nil, errors.New("db error"))

	got, err := svc.Update(context.Background(), "user-1", input)

	assert.Error(t, err)
	assert.Equal(t, "db error", err.Error())
	assert.Nil(t, got)

	repo.AssertExpectations(t)
}
