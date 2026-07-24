package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	user_settings_domain "flatty-budget/go-api/domains/user_settings"
	"flatty-budget/go-api/http/dto"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type mockUserSettingsService struct {
	mock.Mock
}

func (m *mockUserSettingsService) Get(ctx context.Context, userID string) (*user_settings_domain.UserSettings, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*user_settings_domain.UserSettings), args.Error(1)
}

func (m *mockUserSettingsService) Update(ctx context.Context, userID string, input *user_settings_domain.UserSettingsInput) (*user_settings_domain.UserSettings, error) {
	args := m.Called(ctx, userID, input)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*user_settings_domain.UserSettings), args.Error(1)
}

func setupUserSettingsTest() (*gin.Engine, *mockUserSettingsService) {
	gin.SetMode(gin.TestMode)
	mockSvc := new(mockUserSettingsService)
	h := NewUserSettingsHandler(mockSvc)

	r := gin.New()
	r.GET("/user/me/settings", func(c *gin.Context) {
		c.Set("userID", "test-user-id")
		h.GetSettings(c)
	})
	r.PUT("/user/me/settings", func(c *gin.Context) {
		c.Set("userID", "test-user-id")
		h.UpdateSettings(c)
	})

	return r, mockSvc
}

func TestGetSettings_Success(t *testing.T) {
	t.Parallel()

	r, mockSvc := setupUserSettingsTest()

	now := time.Now()
	expected := user_settings_domain.NewUserSettings("test-user-id", "en", "dark", "America/New_York", "MM/DD/YYYY", now, now)

	mockSvc.On("Get", mock.Anything, "test-user-id").Return(expected, nil)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/user/me/settings", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp dto.UserSettingsResponse
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, "test-user-id", resp.UserID)
	assert.Equal(t, "en", resp.Language)
	assert.Equal(t, "dark", resp.Theme)
	assert.Equal(t, "America/New_York", resp.Timezone)
	assert.Equal(t, "MM/DD/YYYY", resp.DateFormat)

	mockSvc.AssertExpectations(t)
}

func TestGetSettings_Defaults(t *testing.T) {
	t.Parallel()

	r, mockSvc := setupUserSettingsTest()

	now := time.Now()
	defaults := user_settings_domain.NewUserSettings("test-user-id", "en", "system", "UTC", "YYYY-MM-DD", now, now)

	mockSvc.On("Get", mock.Anything, "test-user-id").Return(defaults, nil)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/user/me/settings", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp dto.UserSettingsResponse
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, "en", resp.Language)
	assert.Equal(t, "system", resp.Theme)
	assert.Equal(t, "UTC", resp.Timezone)
	assert.Equal(t, "YYYY-MM-DD", resp.DateFormat)

	mockSvc.AssertExpectations(t)
}

func TestGetSettings_InternalError(t *testing.T) {
	t.Parallel()

	r, mockSvc := setupUserSettingsTest()

	mockSvc.On("Get", mock.Anything, "test-user-id").Return(nil, errors.New("db error"))

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/user/me/settings", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)

	var body map[string]string
	err := json.Unmarshal(w.Body.Bytes(), &body)
	assert.NoError(t, err)
	assert.Equal(t, "internal server error", body["error"])

	mockSvc.AssertExpectations(t)
}

func TestUpdateSettings_Success(t *testing.T) {
	t.Parallel()

	r, mockSvc := setupUserSettingsTest()

	now := time.Now()
	theme := "dark"
	reqBody := dto.UpdateUserSettingsRequest{
		Theme: &theme,
	}

	current := user_settings_domain.NewUserSettings("test-user-id", "en", "system", "UTC", "YYYY-MM-DD", now, now)
	mockSvc.On("Get", mock.Anything, "test-user-id").Return(current, nil)

	mergedInput := user_settings_domain.NewUserSettingsInput("en", "dark", "UTC", "YYYY-MM-DD")
	updated := user_settings_domain.NewUserSettings("test-user-id", "en", "dark", "UTC", "YYYY-MM-DD", now, now)
	mockSvc.On("Update", mock.Anything, "test-user-id", mergedInput).Return(updated, nil)

	bodyBytes, _ := json.Marshal(reqBody)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPut, "/user/me/settings", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp dto.UserSettingsResponse
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, "en", resp.Language)
	assert.Equal(t, "dark", resp.Theme)
	assert.Equal(t, "UTC", resp.Timezone)
	assert.Equal(t, "YYYY-MM-DD", resp.DateFormat)

	mockSvc.AssertExpectations(t)
}

func TestUpdateSettings_BadJSON(t *testing.T) {
	t.Parallel()

	r, _ := setupUserSettingsTest()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPut, "/user/me/settings", bytes.NewReader([]byte("{invalid")))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestUpdateSettings_GetError(t *testing.T) {
	t.Parallel()

	r, mockSvc := setupUserSettingsTest()

	theme := "dark"
	reqBody := dto.UpdateUserSettingsRequest{Theme: &theme}

	mockSvc.On("Get", mock.Anything, "test-user-id").Return(nil, errors.New("db error"))

	bodyBytes, _ := json.Marshal(reqBody)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPut, "/user/me/settings", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)

	mockSvc.AssertExpectations(t)
}

func TestUpdateSettings_UpdateError(t *testing.T) {
	t.Parallel()

	r, mockSvc := setupUserSettingsTest()

	now := time.Now()
	theme := "dark"
	reqBody := dto.UpdateUserSettingsRequest{Theme: &theme}

	current := user_settings_domain.NewUserSettings("test-user-id", "en", "system", "UTC", "YYYY-MM-DD", now, now)
	mockSvc.On("Get", mock.Anything, "test-user-id").Return(current, nil)

	mergedInput := user_settings_domain.NewUserSettingsInput("en", "dark", "UTC", "YYYY-MM-DD")
	mockSvc.On("Update", mock.Anything, "test-user-id", mergedInput).Return(nil, errors.New("db error"))

	bodyBytes, _ := json.Marshal(reqBody)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPut, "/user/me/settings", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)

	mockSvc.AssertExpectations(t)
}