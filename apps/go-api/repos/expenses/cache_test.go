package expenses

import (
	"context"
	"errors"
	"testing"
	"time"

	categorydomain "flatty-budget/go-api/domains/category"
	expensesdomain "flatty-budget/go-api/domains/expenses"
	"github.com/stretchr/testify/assert"
)

// ---------------------------------------------------------------------------
// Fake repository
// ---------------------------------------------------------------------------

// fakeRepo implements expensesdomain.Repository in memory, counting calls so
// tests can assert cache hits bypass the wrapped repository.
type fakeRepo struct {
	countCalls       int
	listCalls        int
	createCalls      int
	updateCalls      int
	deleteCalls      int
	getByIDCalls     int
	yearsMonthsCalls int
	byYearMonthCalls int

	countResult   int
	countErr      error
	listResult    []*expensesdomain.ExpenseWithCategory
	listErr       error
	createResult  *expensesdomain.Expense
	createErr     error
	updateResult  *expensesdomain.Expense
	updateErr     error
	deleteResult  int64
	deleteErr     error
	getByIDResult *expensesdomain.Expense
	getByIDErr    error

	yearsMonthsResult []*expensesdomain.YearAndMonth
	yearsMonthsErr    error
	byYearMonthResult []*expensesdomain.ExpenseWithCategory
	byYearMonthErr    error
}

func (f *fakeRepo) Count(ctx context.Context, residentLocationID int64, userID string) (int, error) {
	f.countCalls++
	return f.countResult, f.countErr
}

func (f *fakeRepo) List(ctx context.Context, residentLocationID int64, userID string, limit, offset int) ([]*expensesdomain.ExpenseWithCategory, error) {
	f.listCalls++
	return f.listResult, f.listErr
}

func (f *fakeRepo) Create(ctx context.Context, input *expensesdomain.ExpenseInput, userID string) (*expensesdomain.Expense, error) {
	f.createCalls++
	return f.createResult, f.createErr
}

func (f *fakeRepo) Update(ctx context.Context, id int64, input *expensesdomain.ExpenseInput, userID string) (*expensesdomain.Expense, error) {
	f.updateCalls++
	return f.updateResult, f.updateErr
}

func (f *fakeRepo) Delete(ctx context.Context, id int64, userID string) (int64, error) {
	f.deleteCalls++
	return f.deleteResult, f.deleteErr
}

func (f *fakeRepo) GetByID(ctx context.Context, id int64) (*expensesdomain.Expense, error) {
	f.getByIDCalls++
	return f.getByIDResult, f.getByIDErr
}

func (f *fakeRepo) GetYearsAndMonths(ctx context.Context, residentLocationID int64, userID string) ([]*expensesdomain.YearAndMonth, error) {
	f.yearsMonthsCalls++
	return f.yearsMonthsResult, f.yearsMonthsErr
}

func (f *fakeRepo) GetExpensesByYearMonth(ctx context.Context, residentLocationID, year, month int64, userID string) ([]*expensesdomain.ExpenseWithCategory, error) {
	f.byYearMonthCalls++
	return f.byYearMonthResult, f.byYearMonthErr
}

// ---------------------------------------------------------------------------
// Assertion helpers
// ---------------------------------------------------------------------------

func assertCachedExpenseEqual(t *testing.T, want, got *expensesdomain.Expense) {
	t.Helper()
	assert.Equal(t, want.ID(), got.ID())
	assert.Equal(t, want.ResidentLocationID(), got.ResidentLocationID())
	assert.Equal(t, want.CategoryID(), got.CategoryID())
	assert.Equal(t, want.Amount(), got.Amount())
	assert.Equal(t, want.Month(), got.Month())
	assert.Equal(t, want.Year(), got.Year())
	assert.True(t, want.CreatedAt().Equal(got.CreatedAt()), "CreatedAt mismatch")
	assert.True(t, want.UpdatedAt().Equal(got.UpdatedAt()), "UpdatedAt mismatch")
	assert.Equal(t, want.Description(), got.Description())
	assert.Equal(t, want.UserID(), got.UserID())
}

func assertExpenseWithCategoryEqual(t *testing.T, want, got *expensesdomain.ExpenseWithCategory) {
	t.Helper()
	assert.Equal(t, want.ID(), got.ID())
	assert.Equal(t, want.ResidentLocationID(), got.ResidentLocationID())
	assert.Equal(t, want.CategoryID(), got.CategoryID())
	assert.Equal(t, want.Amount(), got.Amount())
	assert.Equal(t, want.Month(), got.Month())
	assert.Equal(t, want.Year(), got.Year())
	assert.True(t, want.CreatedAt().Equal(got.CreatedAt()), "CreatedAt mismatch")
	assert.True(t, want.UpdatedAt().Equal(got.UpdatedAt()), "UpdatedAt mismatch")
	assert.Equal(t, want.Description(), got.Description())
	assert.Equal(t, want.UserID(), got.UserID())
	wantCat := want.Category()
	gotCat := got.Category()
	assert.Equal(t, wantCat.ID(), gotCat.ID())
	assert.Equal(t, wantCat.Name(), gotCat.Name())
	assert.Equal(t, wantCat.Description(), gotCat.Description())
	assert.True(t, wantCat.CreatedAt().Equal(gotCat.CreatedAt()), "Category CreatedAt mismatch")
	assert.True(t, wantCat.UpdatedAt().Equal(gotCat.UpdatedAt()), "Category UpdatedAt mismatch")
}

func assertCachedExpenseSliceEqual(t *testing.T, want, got []*expensesdomain.ExpenseWithCategory) {
	t.Helper()
	assert.Equal(t, len(want), len(got))
	for i := range want {
		assertExpenseWithCategoryEqual(t, want[i], got[i])
	}
}

func assertYearAndMonthSliceEqual(t *testing.T, want, got []*expensesdomain.YearAndMonth) {
	t.Helper()
	assert.Equal(t, len(want), len(got))
	for i := range want {
		assert.Equal(t, want[i].Year(), got[i].Year())
		assert.Equal(t, want[i].Month(), got[i].Month())
		assert.Equal(t, want[i].Expenses(), got[i].Expenses())
	}
}

func newTestExpenseWithCategory(id, locID, catID int64, amount float64, desc string, month, year int, now time.Time, userID string) *expensesdomain.ExpenseWithCategory {
	return expensesdomain.NewExpenseWithCategory(
		id, locID, catID, amount, desc, month, year, now, now, userID,
		*categorydomain.NewCategory(catID, "Rent", "Monthly rent", now, now),
	)
}

func newTestExpense(id, locID, catID int64, amount float64, desc string, month, year int, now time.Time, userID string) *expensesdomain.Expense {
	return expensesdomain.NewExpense(id, locID, catID, amount, desc, month, year, now, now, userID)
}

func newTestYearAndMonth(year, month, expenses int64) *expensesdomain.YearAndMonth {
	return expensesdomain.NewYearAndMonth(year, month, expenses)
}

// ---------------------------------------------------------------------------
// List
// ---------------------------------------------------------------------------

// TestCachedRepository_List_CacheHit covers both the miss (first call
// populates) and the hit (second call is served from cache — wrapped repo
// untouched).
func TestCachedRepository_List_CacheHit(t *testing.T) {
	t.Parallel()

	now := time.Date(2024, 6, 15, 10, 30, 0, 0, time.UTC)
	items := []*expensesdomain.ExpenseWithCategory{
		newTestExpenseWithCategory(1, 10, 100, 1200.50, "Rent", 6, 2024, now, "user-1"),
	}
	fake := &fakeRepo{listResult: items}
	repo := NewCachedRepository(fake)
	ctx := context.Background()

	first, err := repo.List(ctx, 10, "user-1", 20, 0)
	assert.NoError(t, err)
	assertCachedExpenseSliceEqual(t, items, first)
	assert.Equal(t, 1, fake.listCalls)

	second, err := repo.List(ctx, 10, "user-1", 20, 0)
	assert.NoError(t, err)
	assertCachedExpenseSliceEqual(t, items, second)
	assert.Equal(t, 1, fake.listCalls, "second List with same args must be served from cache")
}

func TestCachedRepository_List_KeyIsolation(t *testing.T) {
	t.Parallel()

	now := time.Date(2024, 6, 15, 10, 30, 0, 0, time.UTC)
	fake := &fakeRepo{listResult: []*expensesdomain.ExpenseWithCategory{
		newTestExpenseWithCategory(1, 10, 100, 1200.50, "Rent", 6, 2024, now, "user-1"),
	}}
	repo := NewCachedRepository(fake)
	ctx := context.Background()

	_, err := repo.List(ctx, 10, "user-1", 20, 0)
	assert.NoError(t, err)
	_, err = repo.List(ctx, 11, "user-1", 20, 0)
	assert.NoError(t, err)
	_, err = repo.List(ctx, 10, "user-2", 20, 0)
	assert.NoError(t, err)
	_, err = repo.List(ctx, 10, "user-1", 10, 0)
	assert.NoError(t, err)
	_, err = repo.List(ctx, 10, "user-1", 20, 0) // repeat: cache hit
	assert.NoError(t, err)

	assert.Equal(t, 4, fake.listCalls, "each distinct (locID, userID, limit, offset) must have its own cache entry")
}

func TestCachedRepository_List_DefensiveCopy(t *testing.T) {
	t.Parallel()

	now := time.Date(2024, 6, 15, 10, 30, 0, 0, time.UTC)
	items := []*expensesdomain.ExpenseWithCategory{
		newTestExpenseWithCategory(1, 10, 100, 1200.50, "Rent", 6, 2024, now, "user-1"),
	}
	fake := &fakeRepo{listResult: items}
	repo := NewCachedRepository(fake)
	ctx := context.Background()

	first, err := repo.List(ctx, 10, "user-1", 20, 0)
	assert.NoError(t, err)
	first = append(first, newTestExpenseWithCategory(99, 10, 100, 1, "Injected", 6, 2024, now, "user-1"))

	second, err := repo.List(ctx, 10, "user-1", 20, 0)
	assert.NoError(t, err)
	assert.Equal(t, 1, len(second), "mutating a returned slice must not corrupt the cached entry")
}

func TestCachedRepository_List_DBErrorNotCached(t *testing.T) {
	t.Parallel()

	now := time.Date(2024, 6, 15, 10, 30, 0, 0, time.UTC)
	items := []*expensesdomain.ExpenseWithCategory{
		newTestExpenseWithCategory(1, 10, 100, 1200.50, "Rent", 6, 2024, now, "user-1"),
	}
	fake := &fakeRepo{listErr: errors.New("db down")}
	repo := NewCachedRepository(fake)
	ctx := context.Background()

	_, err := repo.List(ctx, 10, "user-1", 20, 0)
	assert.Error(t, err)

	fake.listErr = nil
	fake.listResult = items

	got, err := repo.List(ctx, 10, "user-1", 20, 0)
	assert.NoError(t, err)
	assertCachedExpenseSliceEqual(t, items, got)
	assert.Equal(t, 2, fake.listCalls, "DB errors must never be cached")
}

func TestCachedRepository_List_EmptyResultIsCached(t *testing.T) {
	t.Parallel()

	fake := &fakeRepo{listResult: nil}
	repo := NewCachedRepository(fake)
	ctx := context.Background()

	first, err := repo.List(ctx, 10, "user-1", 20, 0)
	assert.NoError(t, err)
	assert.Nil(t, first)
	assert.Equal(t, 1, fake.listCalls)

	second, err := repo.List(ctx, 10, "user-1", 20, 0)
	assert.NoError(t, err)
	assert.Nil(t, second)
	assert.Equal(t, 1, fake.listCalls, "empty results are valid and must be cached")
}

func TestCachedRepository_List_TTLExpiry(t *testing.T) {
	t.Parallel()

	now := time.Date(2024, 6, 15, 10, 30, 0, 0, time.UTC)
	items := []*expensesdomain.ExpenseWithCategory{
		newTestExpenseWithCategory(1, 10, 100, 1200.50, "Rent", 6, 2024, now, "user-1"),
	}
	fake := &fakeRepo{listResult: items}
	repo := NewCachedRepositoryWithTTL(fake, 20*time.Millisecond)
	ctx := context.Background()

	_, err := repo.List(ctx, 10, "user-1", 20, 0)
	assert.NoError(t, err)
	_, err = repo.List(ctx, 10, "user-1", 20, 0)
	assert.NoError(t, err)
	assert.Equal(t, 1, fake.listCalls, "calls within TTL are served from cache")

	time.Sleep(60 * time.Millisecond)

	_, err = repo.List(ctx, 10, "user-1", 20, 0)
	assert.NoError(t, err)
	assert.Equal(t, 2, fake.listCalls, "entry must expire and re-fetch after TTL")
}

// ---------------------------------------------------------------------------
// Count
// ---------------------------------------------------------------------------

func TestCachedRepository_Count_CacheHit(t *testing.T) {
	t.Parallel()

	fake := &fakeRepo{countResult: 42}
	repo := NewCachedRepository(fake)
	ctx := context.Background()

	first, err := repo.Count(ctx, 10, "user-1")
	assert.NoError(t, err)
	assert.Equal(t, 42, first)
	assert.Equal(t, 1, fake.countCalls)

	second, err := repo.Count(ctx, 10, "user-1")
	assert.NoError(t, err)
	assert.Equal(t, 42, second)
	assert.Equal(t, 1, fake.countCalls, "second Count with same args must be served from cache")
}

func TestCachedRepository_Count_DBErrorNotCached(t *testing.T) {
	t.Parallel()

	fake := &fakeRepo{countErr: errors.New("db down")}
	repo := NewCachedRepository(fake)
	ctx := context.Background()

	_, err := repo.Count(ctx, 10, "user-1")
	assert.Error(t, err)

	fake.countErr = nil
	fake.countResult = 7

	got, err := repo.Count(ctx, 10, "user-1")
	assert.NoError(t, err)
	assert.Equal(t, 7, got)
	assert.Equal(t, 2, fake.countCalls, "DB errors must never be cached")
}

// ---------------------------------------------------------------------------
// GetYearsAndMonths
// ---------------------------------------------------------------------------

func TestCachedRepository_GetYearsAndMonths_CacheHit(t *testing.T) {
	t.Parallel()

	items := []*expensesdomain.YearAndMonth{
		newTestYearAndMonth(2024, 6, 3),
		newTestYearAndMonth(2024, 5, 1),
	}
	fake := &fakeRepo{yearsMonthsResult: items}
	repo := NewCachedRepository(fake)
	ctx := context.Background()

	first, err := repo.GetYearsAndMonths(ctx, 10, "user-1")
	assert.NoError(t, err)
	assertYearAndMonthSliceEqual(t, items, first)
	assert.Equal(t, 1, fake.yearsMonthsCalls)

	second, err := repo.GetYearsAndMonths(ctx, 10, "user-1")
	assert.NoError(t, err)
	assertYearAndMonthSliceEqual(t, items, second)
	assert.Equal(t, 1, fake.yearsMonthsCalls, "second GetYearsAndMonths with same args must be served from cache")
}

// ---------------------------------------------------------------------------
// GetExpensesByYearMonth
// ---------------------------------------------------------------------------

func TestCachedRepository_GetExpensesByYearMonth_CacheHit(t *testing.T) {
	t.Parallel()

	now := time.Date(2024, 6, 15, 10, 30, 0, 0, time.UTC)
	items := []*expensesdomain.ExpenseWithCategory{
		newTestExpenseWithCategory(1, 10, 100, 1200.50, "Rent", 6, 2024, now, "user-1"),
	}
	fake := &fakeRepo{byYearMonthResult: items}
	repo := NewCachedRepository(fake)
	ctx := context.Background()

	first, err := repo.GetExpensesByYearMonth(ctx, 10, 2024, 6, "user-1")
	assert.NoError(t, err)
	assertCachedExpenseSliceEqual(t, items, first)
	assert.Equal(t, 1, fake.byYearMonthCalls)

	second, err := repo.GetExpensesByYearMonth(ctx, 10, 2024, 6, "user-1")
	assert.NoError(t, err)
	assertCachedExpenseSliceEqual(t, items, second)
	assert.Equal(t, 1, fake.byYearMonthCalls, "second GetExpensesByYearMonth with same args must be served from cache")
}

func TestCachedRepository_GetExpensesByYearMonth_KeyIsolation(t *testing.T) {
	t.Parallel()

	now := time.Date(2024, 6, 15, 10, 30, 0, 0, time.UTC)
	fake := &fakeRepo{byYearMonthResult: []*expensesdomain.ExpenseWithCategory{
		newTestExpenseWithCategory(1, 10, 100, 1200.50, "Rent", 6, 2024, now, "user-1"),
	}}
	repo := NewCachedRepository(fake)
	ctx := context.Background()

	_, err := repo.GetExpensesByYearMonth(ctx, 10, 2024, 6, "user-1")
	assert.NoError(t, err)
	_, err = repo.GetExpensesByYearMonth(ctx, 10, 2024, 5, "user-1")
	assert.NoError(t, err)
	_, err = repo.GetExpensesByYearMonth(ctx, 10, 2024, 6, "user-2")
	assert.NoError(t, err)
	_, err = repo.GetExpensesByYearMonth(ctx, 10, 2024, 6, "user-1") // repeat: cache hit
	assert.NoError(t, err)

	assert.Equal(t, 3, fake.byYearMonthCalls, "each distinct (year, month, userID) must have its own cache entry")
}

// ---------------------------------------------------------------------------
// GetByID (intentionally not cached)
// ---------------------------------------------------------------------------

// TestCachedRepository_GetByID_NotCached verifies GetByID always reaches the
// wrapped repository — it runs before writes to produce the Kafka pre-image and
// must never return a stale value.
func TestCachedRepository_GetByID_NotCached(t *testing.T) {
	t.Parallel()

	now := time.Date(2024, 6, 15, 10, 30, 0, 0, time.UTC)
	exp := newTestExpense(1, 10, 100, 1200.50, "Rent", 6, 2024, now, "user-1")
	fake := &fakeRepo{getByIDResult: exp}
	repo := NewCachedRepository(fake)
	ctx := context.Background()

	first, err := repo.GetByID(ctx, 1)
	assert.NoError(t, err)
	assertCachedExpenseEqual(t, exp, first)

	second, err := repo.GetByID(ctx, 1)
	assert.NoError(t, err)
	assertCachedExpenseEqual(t, exp, second)

	assert.Equal(t, 2, fake.getByIDCalls, "GetByID must never be cached")
}

// ---------------------------------------------------------------------------
// Invalidation (per-user generation)
// ---------------------------------------------------------------------------

// fullyLoadedRepo returns a fake preloaded so the first List/Count/
// GetYearsAndMonths/GetExpensesByYearMonth calls populate all caches, and a
// CachedRepository over it.
func fullyLoadedRepo(now time.Time) (*fakeRepo, *CachedRepository) {
	fake := &fakeRepo{
		listResult: []*expensesdomain.ExpenseWithCategory{
			newTestExpenseWithCategory(1, 10, 100, 1200.50, "Rent", 6, 2024, now, "user-1"),
		},
		countResult: 1,
		yearsMonthsResult: []*expensesdomain.YearAndMonth{
			newTestYearAndMonth(2024, 6, 1),
		},
		byYearMonthResult: []*expensesdomain.ExpenseWithCategory{
			newTestExpenseWithCategory(1, 10, 100, 1200.50, "Rent", 6, 2024, now, "user-1"),
		},
		createResult: newTestExpense(2, 10, 100, 50, "Groceries", 6, 2024, now, "user-1"),
		updateResult: newTestExpense(1, 10, 100, 999, "Rent", 6, 2024, now, "user-1"),
		deleteResult: 1,
	}
	return fake, NewCachedRepository(fake)
}

func populateAllCaches(t *testing.T, repo *CachedRepository) {
	t.Helper()
	ctx := context.Background()

	_, err := repo.List(ctx, 10, "user-1", 20, 0)
	assert.NoError(t, err)
	_, err = repo.Count(ctx, 10, "user-1")
	assert.NoError(t, err)
	_, err = repo.GetYearsAndMonths(ctx, 10, "user-1")
	assert.NoError(t, err)
	_, err = repo.GetExpensesByYearMonth(ctx, 10, 2024, 6, "user-1")
	assert.NoError(t, err)
}

func TestCachedRepository_InvalidatedByCreate(t *testing.T) {
	t.Parallel()

	now := time.Date(2024, 6, 15, 10, 30, 0, 0, time.UTC)
	fake, repo := fullyLoadedRepo(now)
	ctx := context.Background()
	populateAllCaches(t, repo)
	assert.Equal(t, 1, fake.listCalls)
	assert.Equal(t, 1, fake.countCalls)
	assert.Equal(t, 1, fake.yearsMonthsCalls)
	assert.Equal(t, 1, fake.byYearMonthCalls)

	created, err := repo.Create(ctx, expensesdomain.NewExpenseInput(10, 100, 50, "Groceries", 6, 2024), "user-1")
	assert.NoError(t, err)
	assertCachedExpenseEqual(t, fake.createResult, created)

	_, err = repo.List(ctx, 10, "user-1", 20, 0)
	assert.NoError(t, err)
	_, err = repo.Count(ctx, 10, "user-1")
	assert.NoError(t, err)
	_, err = repo.GetYearsAndMonths(ctx, 10, "user-1")
	assert.NoError(t, err)
	_, err = repo.GetExpensesByYearMonth(ctx, 10, 2024, 6, "user-1")
	assert.NoError(t, err)
	assert.Equal(t, 2, fake.listCalls, "Create must invalidate cached lists")
	assert.Equal(t, 2, fake.countCalls, "Create must invalidate cached counts")
	assert.Equal(t, 2, fake.yearsMonthsCalls, "Create must invalidate cached years-months")
	assert.Equal(t, 2, fake.byYearMonthCalls, "Create must invalidate cached by-year-month")
}

func TestCachedRepository_InvalidatedByUpdate(t *testing.T) {
	t.Parallel()

	now := time.Date(2024, 6, 15, 10, 30, 0, 0, time.UTC)
	fake, repo := fullyLoadedRepo(now)
	ctx := context.Background()
	populateAllCaches(t, repo)

	updated, err := repo.Update(ctx, 1, expensesdomain.NewExpenseInput(10, 100, 999, "Rent", 6, 2024), "user-1")
	assert.NoError(t, err)
	assertCachedExpenseEqual(t, fake.updateResult, updated)

	_, err = repo.List(ctx, 10, "user-1", 20, 0)
	assert.NoError(t, err)
	assert.Equal(t, 2, fake.listCalls, "Update must invalidate cached lists")
}

func TestCachedRepository_InvalidatedByDelete(t *testing.T) {
	t.Parallel()

	now := time.Date(2024, 6, 15, 10, 30, 0, 0, time.UTC)
	fake, repo := fullyLoadedRepo(now)
	ctx := context.Background()
	populateAllCaches(t, repo)

	deleted, err := repo.Delete(ctx, 1, "user-1")
	assert.NoError(t, err)
	assert.Equal(t, int64(1), deleted)

	_, err = repo.List(ctx, 10, "user-1", 20, 0)
	assert.NoError(t, err)
	assert.Equal(t, 2, fake.listCalls, "Delete must invalidate cached lists")
}

// TestCachedRepository_PerUserIsolation verifies a write by one user does NOT
// invalidate another user's cached entries.
func TestCachedRepository_PerUserIsolation(t *testing.T) {
	t.Parallel()

	now := time.Date(2024, 6, 15, 10, 30, 0, 0, time.UTC)
	fake := &fakeRepo{
		listResult: []*expensesdomain.ExpenseWithCategory{
			newTestExpenseWithCategory(1, 10, 100, 1200.50, "Rent", 6, 2024, now, "user-1"),
		},
		createResult: newTestExpense(2, 10, 100, 50, "Groceries", 6, 2024, now, "user-2"),
	}
	repo := NewCachedRepository(fake)
	ctx := context.Background()

	_, err := repo.List(ctx, 10, "user-1", 20, 0)
	assert.NoError(t, err)
	_, err = repo.List(ctx, 10, "user-2", 20, 0)
	assert.NoError(t, err)
	assert.Equal(t, 2, fake.listCalls)

	_, err = repo.Create(ctx, expensesdomain.NewExpenseInput(10, 100, 50, "Groceries", 6, 2024), "user-2")
	assert.NoError(t, err)

	_, err = repo.List(ctx, 10, "user-2", 20, 0)
	assert.NoError(t, err)
	assert.Equal(t, 3, fake.listCalls, "writer's own cache must be invalidated")

	_, err = repo.List(ctx, 10, "user-1", 20, 0)
	assert.NoError(t, err)
	assert.Equal(t, 3, fake.listCalls, "other user's cache must NOT be invalidated")
}

func TestCachedRepository_FailedWriteDoesNotInvalidate(t *testing.T) {
	t.Parallel()

	now := time.Date(2024, 6, 15, 10, 30, 0, 0, time.UTC)
	fake := &fakeRepo{
		listResult: []*expensesdomain.ExpenseWithCategory{
			newTestExpenseWithCategory(1, 10, 100, 1200.50, "Rent", 6, 2024, now, "user-1"),
		},
		createErr: errors.New("insert failed"),
	}
	repo := NewCachedRepository(fake)
	ctx := context.Background()

	_, err := repo.List(ctx, 10, "user-1", 20, 0)
	assert.NoError(t, err)
	assert.Equal(t, 1, fake.listCalls)

	_, err = repo.Create(ctx, expensesdomain.NewExpenseInput(10, 100, 50, "Groceries", 6, 2024), "user-1")
	assert.Error(t, err)

	_, err = repo.List(ctx, 10, "user-1", 20, 0)
	assert.NoError(t, err)
	assert.Equal(t, 1, fake.listCalls, "failed Create must not bump the generation counter")
}

// ---------------------------------------------------------------------------
// Fail-open
// ---------------------------------------------------------------------------

// TestCachedRepository_FailOpen_NilCache verifies that a CachedRepository with
// nil caches (the only realistic failure mode for an in-memory LRU) bypasses
// caching entirely and delegates to the wrapped repository.
func TestCachedRepository_FailOpen_NilCache(t *testing.T) {
	t.Parallel()

	now := time.Date(2024, 6, 15, 10, 30, 0, 0, time.UTC)
	items := []*expensesdomain.ExpenseWithCategory{
		newTestExpenseWithCategory(1, 10, 100, 1200.50, "Rent", 6, 2024, now, "user-1"),
	}
	exp := newTestExpense(1, 10, 100, 1200.50, "Rent", 6, 2024, now, "user-1")
	fake := &fakeRepo{listResult: items, countResult: 1, getByIDResult: exp}
	repo := &CachedRepository{repo: fake} // caches intentionally nil
	ctx := context.Background()

	got, err := repo.List(ctx, 10, "user-1", 20, 0)
	assert.NoError(t, err)
	assertCachedExpenseSliceEqual(t, items, got)
	assert.Equal(t, 1, fake.listCalls)

	count, err := repo.Count(ctx, 10, "user-1")
	assert.NoError(t, err)
	assert.Equal(t, 1, count)
	assert.Equal(t, 1, fake.countCalls)

	expense, err := repo.GetByID(ctx, 1)
	assert.NoError(t, err)
	assertCachedExpenseEqual(t, exp, expense)
	assert.Equal(t, 1, fake.getByIDCalls)
}
