package applications

import (
	"context"
	"testing"
	"time"

	applicationsdomain "flatty-budget/go-api/domains/applications"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type mockRepository struct {
	mock.Mock
}

func (m *mockRepository) ListAll(ctx context.Context) ([]*applicationsdomain.Application, error) {
	args := m.Called(ctx)
	apps, _ := args.Get(0).([]*applicationsdomain.Application)
	return apps, args.Error(1)
}

func (m *mockRepository) ListByEnv(ctx context.Context, env string) ([]*applicationsdomain.Application, error) {
	args := m.Called(ctx, env)
	apps, _ := args.Get(0).([]*applicationsdomain.Application)
	return apps, args.Error(1)
}

func (m *mockRepository) Create(ctx context.Context, input *applicationsdomain.ApplicationInput) (*applicationsdomain.Application, error) {
	args := m.Called(ctx, input)
	app, _ := args.Get(0).(*applicationsdomain.Application)
	return app, args.Error(1)
}

func (m *mockRepository) Update(ctx context.Context, id int64, input *applicationsdomain.ApplicationInput) (*applicationsdomain.Application, error) {
	args := m.Called(ctx, id, input)
	app, _ := args.Get(0).(*applicationsdomain.Application)
	return app, args.Error(1)
}

func (m *mockRepository) Delete(ctx context.Context, id int64) (int64, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(int64), args.Error(1)
}

func TestService_Passthroughs(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	now := time.Date(2026, 7, 31, 10, 0, 0, 0, time.UTC)
	app := applicationsdomain.NewApplication(1, "resident", "development", "resident", "styles",
		"http://localhost:8082", "/external-resident", "/", now, now)
	input := applicationsdomain.NewApplicationInput("resident", "development", "resident", "styles",
		"http://localhost:8082", "/external-resident", "/")

	t.Run("ListByEnv", func(t *testing.T) {
		repo := new(mockRepository)
		repo.On("ListByEnv", ctx, "development").Return([]*applicationsdomain.Application{app}, nil)

		svc := New(repo)
		got, err := svc.ListByEnv(ctx, "development")

		assert.NoError(t, err)
		assert.Len(t, got, 1)
		assert.Equal(t, "resident", got[0].Name())
		repo.AssertExpectations(t)
	})

	t.Run("ListAll", func(t *testing.T) {
		repo := new(mockRepository)
		repo.On("ListAll", ctx).Return([]*applicationsdomain.Application{app}, nil)

		svc := New(repo)
		got, err := svc.ListAll(ctx)

		assert.NoError(t, err)
		assert.Len(t, got, 1)
		assert.Equal(t, "resident", got[0].Name())
		repo.AssertExpectations(t)
	})

	t.Run("Create", func(t *testing.T) {
		repo := new(mockRepository)
		repo.On("Create", ctx, input).Return(app, nil)

		svc := New(repo)
		got, err := svc.Create(ctx, input)

		assert.NoError(t, err)
		assert.Equal(t, int64(1), got.ID())
		repo.AssertExpectations(t)
	})

	t.Run("Update", func(t *testing.T) {
		repo := new(mockRepository)
		repo.On("Update", ctx, int64(1), input).Return(app, nil)

		svc := New(repo)
		got, err := svc.Update(ctx, 1, input)

		assert.NoError(t, err)
		assert.Equal(t, "resident", got.Name())
		repo.AssertExpectations(t)
	})

	t.Run("Delete", func(t *testing.T) {
		repo := new(mockRepository)
		repo.On("Delete", ctx, int64(1)).Return(int64(1), nil)

		svc := New(repo)
		id, err := svc.Delete(ctx, 1)

		assert.NoError(t, err)
		assert.Equal(t, int64(1), id)
		repo.AssertExpectations(t)
	})
}
