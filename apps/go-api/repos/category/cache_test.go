package category

import (
	"context"
	"errors"
	"testing"
	"time"

	categorydomain "flatty-budget/go-api/domains/category"
	"github.com/stretchr/testify/assert"
)

// ---------------------------------------------------------------------------
// Fake repository
// ---------------------------------------------------------------------------

// fakeRepo implements categorydomain.Repository in memory, counting calls so
// tests can assert cache hits bypass the wrapped repository.
type fakeRepo struct {
	listCalls   int
	countCalls  int
	createCalls int
	updateCalls int
	deleteCalls int

	listResult  []*categorydomain.Category
	listErr     error
	countResult int
	countErr    error

	createResult *categorydomain.Category
	createErr    error
	updateResult *categorydomain.Category
	updateErr    error
	deleteResult int64
	deleteErr    error
}

func (f *fakeRepo) Count(ctx context.Context) (int, error) {
	f.countCalls++
	return f.countResult, f.countErr
}

func (f *fakeRepo) List(ctx context.Context, limit, offset int) ([]*categorydomain.Category, error) {
	f.listCalls++
	return f.listResult, f.listErr
}

func (f *fakeRepo) Create(ctx context.Context, input *categorydomain.CategoryInput) (*categorydomain.Category, error) {
	f.createCalls++
	return f.createResult, f.createErr
}

func (f *fakeRepo) Update(ctx context.Context, id int64, input *categorydomain.CategoryInput) (*categorydomain.Category, error) {
	f.updateCalls++
	return f.updateResult, f.updateErr
}

func (f *fakeRepo) Delete(ctx context.Context, id int64) (int64, error) {
	f.deleteCalls++
	return f.deleteResult, f.deleteErr
}

// ---------------------------------------------------------------------------
// TestCachedRepository_List
// ---------------------------------------------------------------------------

// TestCachedRepository_List_CacheHit covers both the miss (first call
// populates) and the hit (second call is served from cache — wrapped repo
// untouched).
func TestCachedRepository_List_CacheHit(t *testing.T) {
	t.Parallel()

	now := time.Date(2024, 6, 15, 10, 30, 0, 0, time.UTC)
	items := []*categorydomain.Category{
		categorydomain.NewCategory(1, "Utilities", "Gas & electricity", now, now),
	}
	fake := &fakeRepo{listResult: items}
	repo := NewCachedRepository(fake)
	ctx := context.Background()

	first, err := repo.List(ctx, 10, 0)
	assert.NoError(t, err)
	assertCategorySliceEqual(t, items, first)
	assert.Equal(t, 1, fake.listCalls)

	second, err := repo.List(ctx, 10, 0)
	assert.NoError(t, err)
	assertCategorySliceEqual(t, items, second)
	assert.Equal(t, 1, fake.listCalls, "second List with same args must be served from cache")
}

func TestCachedRepository_List_KeyIsolation(t *testing.T) {
	t.Parallel()

	now := time.Date(2024, 6, 15, 10, 30, 0, 0, time.UTC)
	fake := &fakeRepo{listResult: []*categorydomain.Category{
		categorydomain.NewCategory(1, "Utilities", "Gas & electricity", now, now),
	}}
	repo := NewCachedRepository(fake)
	ctx := context.Background()

	_, err := repo.List(ctx, 10, 0)
	assert.NoError(t, err)
	_, err = repo.List(ctx, 20, 0)
	assert.NoError(t, err)
	_, err = repo.List(ctx, 10, 10)
	assert.NoError(t, err)
	_, err = repo.List(ctx, 20, 0) // repeat: must be served from cache
	assert.NoError(t, err)

	assert.Equal(t, 3, fake.listCalls, "each distinct (limit, offset) pair must have its own cache entry")
}

func TestCachedRepository_List_DefensiveCopy(t *testing.T) {
	t.Parallel()

	now := time.Date(2024, 6, 15, 10, 30, 0, 0, time.UTC)
	items := []*categorydomain.Category{
		categorydomain.NewCategory(1, "Utilities", "Gas & electricity", now, now),
	}
	fake := &fakeRepo{listResult: items}
	repo := NewCachedRepository(fake)
	ctx := context.Background()

	first, err := repo.List(ctx, 10, 0)
	assert.NoError(t, err)
	first = append(first, categorydomain.NewCategory(99, "Injected", "", now, now))

	second, err := repo.List(ctx, 10, 0)
	assert.NoError(t, err)
	assert.Equal(t, 1, len(second), "mutating a returned slice must not corrupt the cached entry")
}

func TestCachedRepository_List_DBErrorNotCached(t *testing.T) {
	t.Parallel()

	now := time.Date(2024, 6, 15, 10, 30, 0, 0, time.UTC)
	items := []*categorydomain.Category{
		categorydomain.NewCategory(1, "Utilities", "Gas & electricity", now, now),
	}
	fake := &fakeRepo{listErr: errors.New("db down")}
	repo := NewCachedRepository(fake)
	ctx := context.Background()

	_, err := repo.List(ctx, 10, 0)
	assert.Error(t, err)

	fake.listErr = nil
	fake.listResult = items

	got, err := repo.List(ctx, 10, 0)
	assert.NoError(t, err)
	assertCategorySliceEqual(t, items, got)
	assert.Equal(t, 2, fake.listCalls, "DB errors must never be cached")
}

func TestCachedRepository_List_EmptyResultIsCached(t *testing.T) {
	t.Parallel()

	fake := &fakeRepo{listResult: nil}
	repo := NewCachedRepository(fake)
	ctx := context.Background()

	first, err := repo.List(ctx, 10, 0)
	assert.NoError(t, err)
	assert.Nil(t, first)
	assert.Equal(t, 1, fake.listCalls)

	second, err := repo.List(ctx, 10, 0)
	assert.NoError(t, err)
	assert.Nil(t, second)
	assert.Equal(t, 1, fake.listCalls, "empty results are valid and must be cached")
}

func TestCachedRepository_List_TTLExpiry(t *testing.T) {
	t.Parallel()

	now := time.Date(2024, 6, 15, 10, 30, 0, 0, time.UTC)
	items := []*categorydomain.Category{
		categorydomain.NewCategory(1, "Utilities", "Gas & electricity", now, now),
	}
	fake := &fakeRepo{listResult: items}
	repo := NewCachedRepositoryWithTTL(fake, 20*time.Millisecond)
	ctx := context.Background()

	_, err := repo.List(ctx, 10, 0)
	assert.NoError(t, err)
	_, err = repo.List(ctx, 10, 0)
	assert.NoError(t, err)
	assert.Equal(t, 1, fake.listCalls, "calls within TTL are served from cache")

	time.Sleep(60 * time.Millisecond)

	_, err = repo.List(ctx, 10, 0)
	assert.NoError(t, err)
	assert.Equal(t, 2, fake.listCalls, "entry must expire and re-fetch after TTL")
}

// ---------------------------------------------------------------------------
// TestCachedRepository_Count
// ---------------------------------------------------------------------------

func TestCachedRepository_Count_CacheHit(t *testing.T) {
	t.Parallel()

	fake := &fakeRepo{countResult: 42}
	repo := NewCachedRepository(fake)
	ctx := context.Background()

	first, err := repo.Count(ctx)
	assert.NoError(t, err)
	assert.Equal(t, 42, first)
	assert.Equal(t, 1, fake.countCalls)

	second, err := repo.Count(ctx)
	assert.NoError(t, err)
	assert.Equal(t, 42, second)
	assert.Equal(t, 1, fake.countCalls, "second Count must be served from cache")
}

func TestCachedRepository_Count_DBErrorNotCached(t *testing.T) {
	t.Parallel()

	fake := &fakeRepo{countErr: errors.New("db down")}
	repo := NewCachedRepository(fake)
	ctx := context.Background()

	_, err := repo.Count(ctx)
	assert.Error(t, err)

	fake.countErr = nil
	fake.countResult = 7

	got, err := repo.Count(ctx)
	assert.NoError(t, err)
	assert.Equal(t, 7, got)
	assert.Equal(t, 2, fake.countCalls, "DB errors must never be cached")
}

// ---------------------------------------------------------------------------
// Invalidation
// ---------------------------------------------------------------------------

// invalidatingRepo returns a fake preloaded so the first List/Count call
// populates the cache, and a CachedRepository over it.
func invalidatingRepo(now time.Time) (*fakeRepo, *CachedRepository) {
	fake := &fakeRepo{
		listResult: []*categorydomain.Category{
			categorydomain.NewCategory(1, "Utilities", "Gas & electricity", now, now),
		},
		countResult:  1,
		createResult: categorydomain.NewCategory(2, "Rent", "Monthly rent", now, now),
		updateResult: categorydomain.NewCategory(1, "Updated", "Updated desc", now, now),
		deleteResult: 1,
	}
	return fake, NewCachedRepository(fake)
}

func TestCachedRepository_InvalidatedByCreate(t *testing.T) {
	t.Parallel()

	now := time.Date(2024, 6, 15, 10, 30, 0, 0, time.UTC)
	fake, repo := invalidatingRepo(now)
	ctx := context.Background()

	_, err := repo.List(ctx, 10, 0)
	assert.NoError(t, err)
	_, err = repo.Count(ctx)
	assert.NoError(t, err)
	assert.Equal(t, 1, fake.listCalls)
	assert.Equal(t, 1, fake.countCalls)

	created, err := repo.Create(ctx, categorydomain.NewCategoryInput("Rent", "Monthly rent"))
	assert.NoError(t, err)
	assertCategoryEqual(t, fake.createResult, created)

	_, err = repo.List(ctx, 10, 0)
	assert.NoError(t, err)
	_, err = repo.Count(ctx)
	assert.NoError(t, err)
	assert.Equal(t, 2, fake.listCalls, "Create must invalidate cached lists")
	assert.Equal(t, 2, fake.countCalls, "Create must invalidate cached counts")
}

func TestCachedRepository_InvalidatedByUpdate(t *testing.T) {
	t.Parallel()

	now := time.Date(2024, 6, 15, 10, 30, 0, 0, time.UTC)
	fake, repo := invalidatingRepo(now)
	ctx := context.Background()

	_, err := repo.List(ctx, 10, 0)
	assert.NoError(t, err)
	assert.Equal(t, 1, fake.listCalls)

	updated, err := repo.Update(ctx, 1, categorydomain.NewCategoryInput("Updated", "Updated desc"))
	assert.NoError(t, err)
	assertCategoryEqual(t, fake.updateResult, updated)

	_, err = repo.List(ctx, 10, 0)
	assert.NoError(t, err)
	assert.Equal(t, 2, fake.listCalls, "Update must invalidate cached lists")
}

func TestCachedRepository_InvalidatedByDelete(t *testing.T) {
	t.Parallel()

	now := time.Date(2024, 6, 15, 10, 30, 0, 0, time.UTC)
	fake, repo := invalidatingRepo(now)
	ctx := context.Background()

	_, err := repo.List(ctx, 10, 0)
	assert.NoError(t, err)
	assert.Equal(t, 1, fake.listCalls)

	deleted, err := repo.Delete(ctx, 1)
	assert.NoError(t, err)
	assert.Equal(t, int64(1), deleted)

	_, err = repo.List(ctx, 10, 0)
	assert.NoError(t, err)
	assert.Equal(t, 2, fake.listCalls, "Delete must invalidate cached lists")
}

func TestCachedRepository_FailedWriteDoesNotInvalidate(t *testing.T) {
	t.Parallel()

	now := time.Date(2024, 6, 15, 10, 30, 0, 0, time.UTC)
	fake := &fakeRepo{
		listResult: []*categorydomain.Category{
			categorydomain.NewCategory(1, "Utilities", "Gas & electricity", now, now),
		},
		createErr: errors.New("insert failed"),
	}
	repo := NewCachedRepository(fake)
	ctx := context.Background()

	_, err := repo.List(ctx, 10, 0)
	assert.NoError(t, err)
	assert.Equal(t, 1, fake.listCalls)

	_, err = repo.Create(ctx, categorydomain.NewCategoryInput("Rent", "Monthly rent"))
	assert.Error(t, err)

	_, err = repo.List(ctx, 10, 0)
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
	items := []*categorydomain.Category{
		categorydomain.NewCategory(1, "Utilities", "Gas & electricity", now, now),
	}
	fake := &fakeRepo{listResult: items, countResult: 7}
	repo := &CachedRepository{repo: fake} // caches intentionally nil
	ctx := context.Background()

	got, err := repo.List(ctx, 10, 0)
	assert.NoError(t, err)
	assertCategorySliceEqual(t, items, got)
	assert.Equal(t, 1, fake.listCalls)

	count, err := repo.Count(ctx)
	assert.NoError(t, err)
	assert.Equal(t, 7, count)
	assert.Equal(t, 1, fake.countCalls)
}
