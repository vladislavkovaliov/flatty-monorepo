package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	expensedomain "flatty-budget/go-api/domains/expenses"
	expensesservice "flatty-budget/go-api/services/expenses"

	"github.com/gin-gonic/gin"
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

func (m *mockRepo) List(ctx context.Context, residentLocationID int64, userID string, limit, offset int) ([]*expensedomain.Expense, error) {
	args := m.Called(ctx, residentLocationID, userID, limit, offset)
	return args.Get(0).([]*expensedomain.Expense), args.Error(1)
}

func (m *mockRepo) GetByID(ctx context.Context, id int64) (*expensedomain.Expense, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(*expensedomain.Expense), args.Error(1)
}

func (m *mockRepo) Create(ctx context.Context, input *expensedomain.ExpenseInput, userID string) (*expensedomain.Expense, error) {
	args := m.Called(ctx, input, userID)
	return args.Get(0).(*expensedomain.Expense), args.Error(1)
}

func (m *mockRepo) Update(ctx context.Context, id int64, input *expensedomain.ExpenseInput, userID string) (*expensedomain.Expense, error) {
	args := m.Called(ctx, id, input, userID)
	return args.Get(0).(*expensedomain.Expense), args.Error(1)
}

func (m *mockRepo) Delete(ctx context.Context, id int64, userID string) (int64, error) {
	args := m.Called(ctx, id, userID)
	return args.Get(0).(int64), args.Error(1)
}

func newTestHandler(repo expensedomain.Repository) *ExpenseHandler {
	return NewExpenseHandler(expensesservice.New(repo, nil))
}

func newGinContext(w *httptest.ResponseRecorder, method, target string, body []byte, userID string) *gin.Context {
	req := httptest.NewRequest(method, target, bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Set("userID", userID)
	return c
}

func TestExpenseHandler_Count_UsesUserIDFromContext(t *testing.T) {
	gin.SetMode(gin.TestMode)

	repo := new(mockRepo)
	handler := newTestHandler(repo)

	repo.On("Count", mock.Anything, int64(10), "user-123").Return(5, nil)

	w := httptest.NewRecorder()
	c := newGinContext(w, http.MethodGet, "/expenses/count?residentLocationId=10", nil, "user-123")

	handler.Count(c)

	assert.Equal(t, http.StatusOK, w.Code)
	repo.AssertExpectations(t)
}

func TestExpenseHandler_List_UsesUserIDFromContext(t *testing.T) {
	gin.SetMode(gin.TestMode)

	repo := new(mockRepo)
	handler := newTestHandler(repo)

	now := time.Now()
	exp := expensedomain.NewExpense(1, 10, 20, 150.50, "", 6, 2024, now, now, "123456")

	repo.On("List", mock.Anything, int64(10), "user-123", 10, 0).Return([]*expensedomain.Expense{exp}, nil)
	repo.On("Count", mock.Anything, int64(10), "user-123").Return(1, nil)

	w := httptest.NewRecorder()
	c := newGinContext(w, http.MethodGet, "/expenses?residentLocationId=10", nil, "user-123")

	handler.List(c)

	assert.Equal(t, http.StatusOK, w.Code)
	repo.AssertExpectations(t)
}

func TestExpenseHandler_Create_UsesUserIDFromContext(t *testing.T) {
	gin.SetMode(gin.TestMode)

	repo := new(mockRepo)
	handler := newTestHandler(repo)

	now := time.Now()
	created := expensedomain.NewExpense(1, 10, 20, 150.50, "", 6, 2024, now, now, "123456")

	repo.On("Create", mock.Anything, mock.Anything, "user-123").Return(created, nil)

	body, _ := json.Marshal(map[string]any{
		"resident_location_id": 10,
		"category_id":          20,
		"amount":               150.50,
		"month":                6,
		"year":                 2024,
	})

	w := httptest.NewRecorder()
	c := newGinContext(w, http.MethodPost, "/expenses", body, "user-123")

	handler.Create(c)

	assert.Equal(t, http.StatusCreated, w.Code)
	repo.AssertExpectations(t)
}

func TestExpenseHandler_Update_UsesUserIDFromContext(t *testing.T) {
	gin.SetMode(gin.TestMode)

	repo := new(mockRepo)
	handler := newTestHandler(repo)

	now := time.Now()
	prev := expensedomain.NewExpense(1, 10, 20, 100.00, "", 6, 2024, now, now, "123456")
	updated := expensedomain.NewExpense(1, 10, 20, 200.00, "", 6, 2024, now, now, "123456")

	repo.On("GetByID", mock.Anything, int64(1)).Return(prev, nil)
	repo.On("Update", mock.Anything, int64(1), mock.Anything, "user-123").Return(updated, nil)

	body, _ := json.Marshal(map[string]any{
		"resident_location_id": 10,
		"category_id":          20,
		"amount":               200.00,
		"month":                6,
		"year":                 2024,
	})

	w := httptest.NewRecorder()
	c := newGinContext(w, http.MethodPut, "/expenses/1", body, "user-123")
	c.Params = gin.Params{{Key: "id", Value: "1"}}

	handler.Update(c)

	assert.Equal(t, http.StatusCreated, w.Code)
	repo.AssertExpectations(t)
}

func TestExpenseHandler_Delete_UsesUserIDFromContext(t *testing.T) {
	gin.SetMode(gin.TestMode)

	repo := new(mockRepo)
	handler := newTestHandler(repo)

	now := time.Now()
	prev := expensedomain.NewExpense(1, 10, 20, 100.00, "", 6, 2024, now, now, "123456")

	repo.On("GetByID", mock.Anything, int64(1)).Return(prev, nil)
	repo.On("Delete", mock.Anything, int64(1), "user-123").Return(int64(1), nil)

	w := httptest.NewRecorder()
	c := newGinContext(w, http.MethodDelete, "/expenses/1", nil, "user-123")
	c.Params = gin.Params{{Key: "id", Value: "1"}}

	handler.Delete(c)

	assert.Equal(t, http.StatusOK, w.Code)
	repo.AssertExpectations(t)
}
