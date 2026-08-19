package expense_stats

import (
	"context"
	"errors"
	"testing"
	"time"

	expensestatsdomain "flatty-budget/go-api/domains/expense_stats"
	"github.com/stretchr/testify/assert"
)

// ---------------------------------------------------------------------------
// Fake repositories
// ---------------------------------------------------------------------------

// fakeTotalRepo implements expensestatsdomain.MonthlyTotalRepository in memory,
// counting calls so tests can assert cache hits bypass the wrapped repository.
type fakeTotalRepo struct {
	listCalls   int
	upsertCalls int

	listResult []*expensestatsdomain.ExpenseMonthlyTotal
	listErr    error
	upsertErr  error
}

func (f *fakeTotalRepo) List(ctx context.Context, residentLocationID int64, userID string, month, year *int) ([]*expensestatsdomain.ExpenseMonthlyTotal, error) {
	f.listCalls++
	return f.listResult, f.listErr
}

func (f *fakeTotalRepo) UpsertTotal(ctx context.Context, residentLocationID int64, month, year int, totalSpent float64) error {
	f.upsertCalls++
	return f.upsertErr
}

// fakeAvgRepo implements expensestatsdomain.MonthlyAverageRepository in memory,
// counting calls so tests can assert cache hits bypass the wrapped repository.
type fakeAvgRepo struct {
	listCalls   int
	upsertCalls int

	listResult []*expensestatsdomain.ExpenseMonthlyAverage
	listErr    error
	upsertErr  error
}

func (f *fakeAvgRepo) List(ctx context.Context, residentLocationID int64, userID string, month, year *int) ([]*expensestatsdomain.ExpenseMonthlyAverage, error) {
	f.listCalls++
	return f.listResult, f.listErr
}

func (f *fakeAvgRepo) UpsertAverage(ctx context.Context, residentLocationID int64, month, year int, averageAmount float64, expenseCount int) error {
	f.upsertCalls++
	return f.upsertErr
}

// ---------------------------------------------------------------------------
// Assertion helpers
// ---------------------------------------------------------------------------

func assertCachedMonthlyTotalEqual(t *testing.T, want, got *expensestatsdomain.ExpenseMonthlyTotal) {
	t.Helper()
	assert.Equal(t, want.ResidentLocationID(), got.ResidentLocationID())
	assert.Equal(t, want.Month(), got.Month())
	assert.Equal(t, want.Year(), got.Year())
	assert.Equal(t, want.TotalSpent(), got.TotalSpent())
}

func assertCachedMonthlyTotalSliceEqual(t *testing.T, want, got []*expensestatsdomain.ExpenseMonthlyTotal) {
	t.Helper()
	assert.Equal(t, len(want), len(got))
	for i := range want {
		assertCachedMonthlyTotalEqual(t, want[i], got[i])
	}
}

func assertCachedMonthlyAverageEqual(t *testing.T, want, got *expensestatsdomain.ExpenseMonthlyAverage) {
	t.Helper()
	assert.Equal(t, want.ResidentLocationID(), got.ResidentLocationID())
	assert.Equal(t, want.Month(), got.Month())
	assert.Equal(t, want.Year(), got.Year())
	assert.Equal(t, want.AverageAmount(), got.AverageAmount())
	assert.Equal(t, want.ExpenseCount(), got.ExpenseCount())
}

func assertCachedMonthlyAverageSliceEqual(t *testing.T, want, got []*expensestatsdomain.ExpenseMonthlyAverage) {
	t.Helper()
	assert.Equal(t, len(want), len(got))
	for i := range want {
		assertCachedMonthlyAverageEqual(t, want[i], got[i])
	}
}

func newTestMonthlyTotal(locID int64, month, year int, total float64) *expensestatsdomain.ExpenseMonthlyTotal {
	return expensestatsdomain.NewExpenseMonthlyTotal(locID, month, year, total)
}

func newTestMonthlyAverage(locID int64, month, year int, avg float64, count int) *expensestatsdomain.ExpenseMonthlyAverage {
	return expensestatsdomain.NewExpenseMonthlyAverage(locID, month, year, avg, count)
}

// ---------------------------------------------------------------------------
// CachedMonthlyTotalRepository
// ---------------------------------------------------------------------------

// TestCachedMonthlyTotalRepository_List_CacheHit covers both the miss (first
// call populates) and the hit (second call is served from cache — wrapped repo
// untouched).
func TestCachedMonthlyTotalRepository_List_CacheHit(t *testing.T) {
	t.Parallel()

	items := []*expensestatsdomain.ExpenseMonthlyTotal{
		newTestMonthlyTotal(10, 6, 2024, 1200.50),
	}
	fake := &fakeTotalRepo{listResult: items}
	repo := NewCachedMonthlyTotalRepository(fake)
	ctx := context.Background()

	first, err := repo.List(ctx, 10, "user-1", intPtr(6), intPtr(2024))
	assert.NoError(t, err)
	assertCachedMonthlyTotalSliceEqual(t, items, first)
	assert.Equal(t, 1, fake.listCalls)

	second, err := repo.List(ctx, 10, "user-1", intPtr(6), intPtr(2024))
	assert.NoError(t, err)
	assertCachedMonthlyTotalSliceEqual(t, items, second)
	assert.Equal(t, 1, fake.listCalls, "second List with same args must be served from cache")
}

// TestCachedMonthlyTotalRepository_List_KeyIsolation verifies nil and non-nil
// month/year filters get distinct cache entries.
func TestCachedMonthlyTotalRepository_List_KeyIsolation(t *testing.T) {
	t.Parallel()

	fake := &fakeTotalRepo{listResult: []*expensestatsdomain.ExpenseMonthlyTotal{
		newTestMonthlyTotal(10, 6, 2024, 1200.50),
	}}
	repo := NewCachedMonthlyTotalRepository(fake)
	ctx := context.Background()

	_, err := repo.List(ctx, 10, "user-1", nil, nil)
	assert.NoError(t, err)
	_, err = repo.List(ctx, 10, "user-1", intPtr(6), intPtr(2024))
	assert.NoError(t, err)
	_, err = repo.List(ctx, 11, "user-1", nil, nil)
	assert.NoError(t, err)
	_, err = repo.List(ctx, 10, "user-2", nil, nil)
	assert.NoError(t, err)
	_, err = repo.List(ctx, 10, "user-1", nil, nil) // repeat: cache hit
	assert.NoError(t, err)

	assert.Equal(t, 4, fake.listCalls, "each distinct (locID, userID, month, year) must have its own cache entry")
}

func TestCachedMonthlyTotalRepository_List_DefensiveCopy(t *testing.T) {
	t.Parallel()

	items := []*expensestatsdomain.ExpenseMonthlyTotal{
		newTestMonthlyTotal(10, 6, 2024, 1200.50),
	}
	fake := &fakeTotalRepo{listResult: items}
	repo := NewCachedMonthlyTotalRepository(fake)
	ctx := context.Background()

	first, err := repo.List(ctx, 10, "user-1", nil, nil)
	assert.NoError(t, err)
	first = append(first, newTestMonthlyTotal(10, 7, 2024, 1))

	second, err := repo.List(ctx, 10, "user-1", nil, nil)
	assert.NoError(t, err)
	assert.Equal(t, 1, len(second), "mutating a returned slice must not corrupt the cached entry")
}

func TestCachedMonthlyTotalRepository_List_DBErrorNotCached(t *testing.T) {
	t.Parallel()

	items := []*expensestatsdomain.ExpenseMonthlyTotal{
		newTestMonthlyTotal(10, 6, 2024, 1200.50),
	}
	fake := &fakeTotalRepo{listErr: errors.New("db down")}
	repo := NewCachedMonthlyTotalRepository(fake)
	ctx := context.Background()

	_, err := repo.List(ctx, 10, "user-1", nil, nil)
	assert.Error(t, err)

	fake.listErr = nil
	fake.listResult = items

	got, err := repo.List(ctx, 10, "user-1", nil, nil)
	assert.NoError(t, err)
	assertCachedMonthlyTotalSliceEqual(t, items, got)
	assert.Equal(t, 2, fake.listCalls, "DB errors must never be cached")
}

func TestCachedMonthlyTotalRepository_List_EmptyResultIsCached(t *testing.T) {
	t.Parallel()

	fake := &fakeTotalRepo{listResult: nil}
	repo := NewCachedMonthlyTotalRepository(fake)
	ctx := context.Background()

	first, err := repo.List(ctx, 10, "user-1", nil, nil)
	assert.NoError(t, err)
	assert.Nil(t, first)
	assert.Equal(t, 1, fake.listCalls)

	second, err := repo.List(ctx, 10, "user-1", nil, nil)
	assert.NoError(t, err)
	assert.Nil(t, second)
	assert.Equal(t, 1, fake.listCalls, "empty results are valid and must be cached")
}

func TestCachedMonthlyTotalRepository_List_TTLExpiry(t *testing.T) {
	t.Parallel()

	items := []*expensestatsdomain.ExpenseMonthlyTotal{
		newTestMonthlyTotal(10, 6, 2024, 1200.50),
	}
	fake := &fakeTotalRepo{listResult: items}
	repo := NewCachedMonthlyTotalRepositoryWithTTL(fake, 20*time.Millisecond)
	ctx := context.Background()

	_, err := repo.List(ctx, 10, "user-1", nil, nil)
	assert.NoError(t, err)
	_, err = repo.List(ctx, 10, "user-1", nil, nil)
	assert.NoError(t, err)
	assert.Equal(t, 1, fake.listCalls, "calls within TTL are served from cache")

	time.Sleep(60 * time.Millisecond)

	_, err = repo.List(ctx, 10, "user-1", nil, nil)
	assert.NoError(t, err)
	assert.Equal(t, 2, fake.listCalls, "entry must expire and re-fetch after TTL")
}

// TestCachedMonthlyTotalRepository_UpsertTotal_PassThrough verifies UpsertTotal
// does NOT invalidate the cache: production writes arrive via the Kafka
// consumer (not through this repo), so staleness is bounded by the TTL alone.
func TestCachedMonthlyTotalRepository_UpsertTotal_PassThrough(t *testing.T) {
	t.Parallel()

	items := []*expensestatsdomain.ExpenseMonthlyTotal{
		newTestMonthlyTotal(10, 6, 2024, 1200.50),
	}
	fake := &fakeTotalRepo{listResult: items}
	repo := NewCachedMonthlyTotalRepository(fake)
	ctx := context.Background()

	_, err := repo.List(ctx, 10, "user-1", nil, nil)
	assert.NoError(t, err)
	assert.Equal(t, 1, fake.listCalls)

	err = repo.UpsertTotal(ctx, 10, 6, 2024, 999)
	assert.NoError(t, err)
	assert.Equal(t, 1, fake.upsertCalls)

	_, err = repo.List(ctx, 10, "user-1", nil, nil)
	assert.NoError(t, err)
	assert.Equal(t, 1, fake.listCalls, "UpsertTotal is a pass-through and must NOT invalidate the cache")
}

// TestCachedMonthlyTotalRepository_FailOpen_NilCache verifies that a decorator
// with a nil cache bypasses caching entirely and delegates to the wrapped
// repository.
func TestCachedMonthlyTotalRepository_FailOpen_NilCache(t *testing.T) {
	t.Parallel()

	items := []*expensestatsdomain.ExpenseMonthlyTotal{
		newTestMonthlyTotal(10, 6, 2024, 1200.50),
	}
	fake := &fakeTotalRepo{listResult: items}
	repo := &CachedMonthlyTotalRepository{repo: fake} // cache intentionally nil
	ctx := context.Background()

	got, err := repo.List(ctx, 10, "user-1", nil, nil)
	assert.NoError(t, err)
	assertCachedMonthlyTotalSliceEqual(t, items, got)
	assert.Equal(t, 1, fake.listCalls)
}

// ---------------------------------------------------------------------------
// CachedMonthlyAverageRepository
// ---------------------------------------------------------------------------

func TestCachedMonthlyAverageRepository_List_CacheHit(t *testing.T) {
	t.Parallel()

	items := []*expensestatsdomain.ExpenseMonthlyAverage{
		newTestMonthlyAverage(10, 6, 2024, 400.25, 3),
	}
	fake := &fakeAvgRepo{listResult: items}
	repo := NewCachedMonthlyAverageRepository(fake)
	ctx := context.Background()

	first, err := repo.List(ctx, 10, "user-1", intPtr(6), intPtr(2024))
	assert.NoError(t, err)
	assertCachedMonthlyAverageSliceEqual(t, items, first)
	assert.Equal(t, 1, fake.listCalls)

	second, err := repo.List(ctx, 10, "user-1", intPtr(6), intPtr(2024))
	assert.NoError(t, err)
	assertCachedMonthlyAverageSliceEqual(t, items, second)
	assert.Equal(t, 1, fake.listCalls, "second List with same args must be served from cache")
}

func TestCachedMonthlyAverageRepository_List_KeyIsolation(t *testing.T) {
	t.Parallel()

	fake := &fakeAvgRepo{listResult: []*expensestatsdomain.ExpenseMonthlyAverage{
		newTestMonthlyAverage(10, 6, 2024, 400.25, 3),
	}}
	repo := NewCachedMonthlyAverageRepository(fake)
	ctx := context.Background()

	_, err := repo.List(ctx, 10, "user-1", nil, nil)
	assert.NoError(t, err)
	_, err = repo.List(ctx, 10, "user-1", intPtr(6), intPtr(2024))
	assert.NoError(t, err)
	_, err = repo.List(ctx, 10, "user-1", nil, nil) // repeat: cache hit
	assert.NoError(t, err)

	assert.Equal(t, 2, fake.listCalls, "each distinct (locID, userID, month, year) must have its own cache entry")
}

func TestCachedMonthlyAverageRepository_List_DBErrorNotCached(t *testing.T) {
	t.Parallel()

	items := []*expensestatsdomain.ExpenseMonthlyAverage{
		newTestMonthlyAverage(10, 6, 2024, 400.25, 3),
	}
	fake := &fakeAvgRepo{listErr: errors.New("db down")}
	repo := NewCachedMonthlyAverageRepository(fake)
	ctx := context.Background()

	_, err := repo.List(ctx, 10, "user-1", nil, nil)
	assert.Error(t, err)

	fake.listErr = nil
	fake.listResult = items

	got, err := repo.List(ctx, 10, "user-1", nil, nil)
	assert.NoError(t, err)
	assertCachedMonthlyAverageSliceEqual(t, items, got)
	assert.Equal(t, 2, fake.listCalls, "DB errors must never be cached")
}

// TestCachedMonthlyAverageRepository_UpsertAverage_PassThrough verifies
// UpsertAverage does NOT invalidate the cache (same Kafka-consumer rationale as
// UpsertTotal).
func TestCachedMonthlyAverageRepository_UpsertAverage_PassThrough(t *testing.T) {
	t.Parallel()

	items := []*expensestatsdomain.ExpenseMonthlyAverage{
		newTestMonthlyAverage(10, 6, 2024, 400.25, 3),
	}
	fake := &fakeAvgRepo{listResult: items}
	repo := NewCachedMonthlyAverageRepository(fake)
	ctx := context.Background()

	_, err := repo.List(ctx, 10, "user-1", nil, nil)
	assert.NoError(t, err)
	assert.Equal(t, 1, fake.listCalls)

	err = repo.UpsertAverage(ctx, 10, 6, 2024, 999, 1)
	assert.NoError(t, err)
	assert.Equal(t, 1, fake.upsertCalls)

	_, err = repo.List(ctx, 10, "user-1", nil, nil)
	assert.NoError(t, err)
	assert.Equal(t, 1, fake.listCalls, "UpsertAverage is a pass-through and must NOT invalidate the cache")
}

func TestCachedMonthlyAverageRepository_FailOpen_NilCache(t *testing.T) {
	t.Parallel()

	items := []*expensestatsdomain.ExpenseMonthlyAverage{
		newTestMonthlyAverage(10, 6, 2024, 400.25, 3),
	}
	fake := &fakeAvgRepo{listResult: items}
	repo := &CachedMonthlyAverageRepository{repo: fake} // cache intentionally nil
	ctx := context.Background()

	got, err := repo.List(ctx, 10, "user-1", nil, nil)
	assert.NoError(t, err)
	assertCachedMonthlyAverageSliceEqual(t, items, got)
	assert.Equal(t, 1, fake.listCalls)
}
