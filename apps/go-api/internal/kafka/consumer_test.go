package kafka

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// mockStatsUpdater implements StatsUpdater.
type mockStatsUpdater struct {
	mock.Mock
}

func (m *mockStatsUpdater) UpsertTotal(ctx context.Context, residentLocationID int64, month, year int, totalSpent float64) error {
	args := m.Called(ctx, residentLocationID, month, year, totalSpent)
	return args.Error(0)
}

func (m *mockStatsUpdater) UpsertAverage(ctx context.Context, residentLocationID int64, month, year int, averageAmount float64, expenseCount int) error {
	args := m.Called(ctx, residentLocationID, month, year, averageAmount, expenseCount)
	return args.Error(0)
}

func (m *mockStatsUpdater) GetTotal(ctx context.Context, residentLocationID int64, month, year int) (float64, error) {
	args := m.Called(ctx, residentLocationID, month, year)
	return args.Get(0).(float64), args.Error(1)
}

func (m *mockStatsUpdater) GetAverage(ctx context.Context, residentLocationID int64, month, year int) (float64, int, error) {
	args := m.Called(ctx, residentLocationID, month, year)
	return args.Get(0).(float64), args.Get(1).(int), args.Error(2)
}

func TestConsumer_HandleCreated_ScopesByResidentLocation(t *testing.T) {
	t.Parallel()

	stats := new(mockStatsUpdater)
	c := &Consumer{stats: stats}

	event := ExpenseEvent{
		Action:             "created",
		ID:                 1,
		ResidentLocationID: 42,
		Month:              6,
		Year:               2026,
		Amount:             100.0,
	}

	stats.On("GetTotal", mock.Anything, int64(42), 6, 2026).Return(0.0, nil)
	stats.On("UpsertTotal", mock.Anything, int64(42), 6, 2026, 100.0).Return(nil)
	stats.On("GetAverage", mock.Anything, int64(42), 6, 2026).Return(0.0, 0, nil)
	stats.On("UpsertAverage", mock.Anything, int64(42), 6, 2026, 100.0, 1).Return(nil)

	err := c.processEvent(context.Background(), event)

	assert.NoError(t, err)
	stats.AssertExpectations(t)
}

func TestConsumer_HandleCreated_DoesNotLeakAcrossResidentLocations(t *testing.T) {
	t.Parallel()

	stats := new(mockStatsUpdater)
	c := &Consumer{stats: stats}

	eventA := ExpenseEvent{Action: "created", ResidentLocationID: 1, Month: 6, Year: 2026, Amount: 50.0}
	eventB := ExpenseEvent{Action: "created", ResidentLocationID: 2, Month: 6, Year: 2026, Amount: 75.0}

	stats.On("GetTotal", mock.Anything, int64(1), 6, 2026).Return(0.0, nil)
	stats.On("UpsertTotal", mock.Anything, int64(1), 6, 2026, 50.0).Return(nil)
	stats.On("GetAverage", mock.Anything, int64(1), 6, 2026).Return(0.0, 0, nil)
	stats.On("UpsertAverage", mock.Anything, int64(1), 6, 2026, 50.0, 1).Return(nil)

	stats.On("GetTotal", mock.Anything, int64(2), 6, 2026).Return(0.0, nil)
	stats.On("UpsertTotal", mock.Anything, int64(2), 6, 2026, 75.0).Return(nil)
	stats.On("GetAverage", mock.Anything, int64(2), 6, 2026).Return(0.0, 0, nil)
	stats.On("UpsertAverage", mock.Anything, int64(2), 6, 2026, 75.0, 1).Return(nil)

	assert.NoError(t, c.processEvent(context.Background(), eventA))
	assert.NoError(t, c.processEvent(context.Background(), eventB))

	// Each resident location's total/average was upserted independently,
	// never combined into a single shared (month, year) bucket.
	stats.AssertExpectations(t)
}
