package residentlocation

import (
	"context"
	"errors"
	"testing"
	"time"

	residentlocationdomain "flatty-budget/go-api/domains/resident_location"
	"github.com/stretchr/testify/assert"
)

// ---------------------------------------------------------------------------
// Fake repository
// ---------------------------------------------------------------------------

// fakeRepo implements residentlocationdomain.Repository in memory, counting
// calls so tests can assert cache hits bypass the wrapped repository.
type fakeRepo struct {
	countCalls  int
	listCalls   int
	createCalls int
	updateCalls int
	deleteCalls int

	countResult  int
	countErr     error
	listResult   []*residentlocationdomain.ResidentLocation
	listErr      error
	createResult *residentlocationdomain.ResidentLocation
	createErr    error
	updateResult *residentlocationdomain.ResidentLocation
	updateErr    error
	deleteResult int64
	deleteErr    error
}

func (f *fakeRepo) Count(ctx context.Context, userID string) (int, error) {
	f.countCalls++
	return f.countResult, f.countErr
}

func (f *fakeRepo) List(ctx context.Context, limit, offset int, userID string) ([]*residentlocationdomain.ResidentLocation, error) {
	f.listCalls++
	return f.listResult, f.listErr
}

func (f *fakeRepo) Create(ctx context.Context, input *residentlocationdomain.ResidentLocationInput, userID string) (*residentlocationdomain.ResidentLocation, error) {
	f.createCalls++
	return f.createResult, f.createErr
}

func (f *fakeRepo) Update(ctx context.Context, id int64, input *residentlocationdomain.ResidentLocationInput, userID string) (*residentlocationdomain.ResidentLocation, error) {
	f.updateCalls++
	return f.updateResult, f.updateErr
}

func (f *fakeRepo) Delete(ctx context.Context, id int64, userID string) (int64, error) {
	f.deleteCalls++
	return f.deleteResult, f.deleteErr
}

// ---------------------------------------------------------------------------
// Assertion helpers
// ---------------------------------------------------------------------------

func assertCachedResidentLocationEqual(t *testing.T, want, got *residentlocationdomain.ResidentLocation) {
	t.Helper()
	assert.Equal(t, want.ID(), got.ID())
	assert.Equal(t, want.UserID(), got.UserID())
	assert.Equal(t, want.Country(), got.Country())
	assert.Equal(t, want.City(), got.City())
	assert.Equal(t, want.PostalCode(), got.PostalCode())
	assert.Equal(t, want.Street(), got.Street())
	assert.Equal(t, want.House(), got.House())
	assert.Equal(t, want.Apartment(), got.Apartment())
	assert.True(t, want.CreatedAt().Equal(got.CreatedAt()), "CreatedAt mismatch")
	assert.True(t, want.UpdatedAt().Equal(got.UpdatedAt()), "UpdatedAt mismatch")
}

func assertCachedResidentLocationSliceEqual(t *testing.T, want, got []*residentlocationdomain.ResidentLocation) {
	t.Helper()
	assert.Equal(t, len(want), len(got))
	for i := range want {
		assertCachedResidentLocationEqual(t, want[i], got[i])
	}
}

func newTestLocation(id int64, userID string, now time.Time) *residentlocationdomain.ResidentLocation {
	return residentlocationdomain.NewResidentLocation(
		id, userID, "Belarus", "Minsk", "220000", "Nezavisimosti", "10", "5",
		now, now,
	)
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
	items := []*residentlocationdomain.ResidentLocation{newTestLocation(1, "user-1", now)}
	fake := &fakeRepo{listResult: items}
	repo := NewCachedRepository(fake)
	ctx := context.Background()

	first, err := repo.List(ctx, 20, 0, "user-1")
	assert.NoError(t, err)
	assertCachedResidentLocationSliceEqual(t, items, first)
	assert.Equal(t, 1, fake.listCalls)

	second, err := repo.List(ctx, 20, 0, "user-1")
	assert.NoError(t, err)
	assertCachedResidentLocationSliceEqual(t, items, second)
	assert.Equal(t, 1, fake.listCalls, "second List with same args must be served from cache")
}

func TestCachedRepository_List_KeyIsolation(t *testing.T) {
	t.Parallel()

	now := time.Date(2024, 6, 15, 10, 30, 0, 0, time.UTC)
	fake := &fakeRepo{listResult: []*residentlocationdomain.ResidentLocation{newTestLocation(1, "user-1", now)}}
	repo := NewCachedRepository(fake)
	ctx := context.Background()

	_, err := repo.List(ctx, 20, 0, "user-1")
	assert.NoError(t, err)
	_, err = repo.List(ctx, 10, 0, "user-1")
	assert.NoError(t, err)
	_, err = repo.List(ctx, 20, 0, "user-2")
	assert.NoError(t, err)
	_, err = repo.List(ctx, 20, 0, "user-1") // repeat: cache hit
	assert.NoError(t, err)

	assert.Equal(t, 3, fake.listCalls, "each distinct (limit, offset, userID) must have its own cache entry")
}

func TestCachedRepository_List_DefensiveCopy(t *testing.T) {
	t.Parallel()

	now := time.Date(2024, 6, 15, 10, 30, 0, 0, time.UTC)
	items := []*residentlocationdomain.ResidentLocation{newTestLocation(1, "user-1", now)}
	fake := &fakeRepo{listResult: items}
	repo := NewCachedRepository(fake)
	ctx := context.Background()

	first, err := repo.List(ctx, 20, 0, "user-1")
	assert.NoError(t, err)
	first = append(first, newTestLocation(99, "user-1", now))

	second, err := repo.List(ctx, 20, 0, "user-1")
	assert.NoError(t, err)
	assert.Equal(t, 1, len(second), "mutating a returned slice must not corrupt the cached entry")
}

func TestCachedRepository_List_DBErrorNotCached(t *testing.T) {
	t.Parallel()

	now := time.Date(2024, 6, 15, 10, 30, 0, 0, time.UTC)
	items := []*residentlocationdomain.ResidentLocation{newTestLocation(1, "user-1", now)}
	fake := &fakeRepo{listErr: errors.New("db down")}
	repo := NewCachedRepository(fake)
	ctx := context.Background()

	_, err := repo.List(ctx, 20, 0, "user-1")
	assert.Error(t, err)

	fake.listErr = nil
	fake.listResult = items

	got, err := repo.List(ctx, 20, 0, "user-1")
	assert.NoError(t, err)
	assertCachedResidentLocationSliceEqual(t, items, got)
	assert.Equal(t, 2, fake.listCalls, "DB errors must never be cached")
}

func TestCachedRepository_List_EmptyResultIsCached(t *testing.T) {
	t.Parallel()

	fake := &fakeRepo{listResult: nil}
	repo := NewCachedRepository(fake)
	ctx := context.Background()

	first, err := repo.List(ctx, 20, 0, "user-1")
	assert.NoError(t, err)
	assert.Nil(t, first)
	assert.Equal(t, 1, fake.listCalls)

	second, err := repo.List(ctx, 20, 0, "user-1")
	assert.NoError(t, err)
	assert.Nil(t, second)
	assert.Equal(t, 1, fake.listCalls, "empty results are valid and must be cached")
}

func TestCachedRepository_List_TTLExpiry(t *testing.T) {
	t.Parallel()

	now := time.Date(2024, 6, 15, 10, 30, 0, 0, time.UTC)
	items := []*residentlocationdomain.ResidentLocation{newTestLocation(1, "user-1", now)}
	fake := &fakeRepo{listResult: items}
	repo := NewCachedRepositoryWithTTL(fake, 20*time.Millisecond)
	ctx := context.Background()

	_, err := repo.List(ctx, 20, 0, "user-1")
	assert.NoError(t, err)
	_, err = repo.List(ctx, 20, 0, "user-1")
	assert.NoError(t, err)
	assert.Equal(t, 1, fake.listCalls, "calls within TTL are served from cache")

	time.Sleep(60 * time.Millisecond)

	_, err = repo.List(ctx, 20, 0, "user-1")
	assert.NoError(t, err)
	assert.Equal(t, 2, fake.listCalls, "entry must expire and re-fetch after TTL")
}

// ---------------------------------------------------------------------------
// Count
// ---------------------------------------------------------------------------

func TestCachedRepository_Count_CacheHit(t *testing.T) {
	t.Parallel()

	fake := &fakeRepo{countResult: 3}
	repo := NewCachedRepository(fake)
	ctx := context.Background()

	first, err := repo.Count(ctx, "user-1")
	assert.NoError(t, err)
	assert.Equal(t, 3, first)
	assert.Equal(t, 1, fake.countCalls)

	second, err := repo.Count(ctx, "user-1")
	assert.NoError(t, err)
	assert.Equal(t, 3, second)
	assert.Equal(t, 1, fake.countCalls, "second Count with same args must be served from cache")
}

func TestCachedRepository_Count_DBErrorNotCached(t *testing.T) {
	t.Parallel()

	fake := &fakeRepo{countErr: errors.New("db down")}
	repo := NewCachedRepository(fake)
	ctx := context.Background()

	_, err := repo.Count(ctx, "user-1")
	assert.Error(t, err)

	fake.countErr = nil
	fake.countResult = 5

	got, err := repo.Count(ctx, "user-1")
	assert.NoError(t, err)
	assert.Equal(t, 5, got)
	assert.Equal(t, 2, fake.countCalls, "DB errors must never be cached")
}

// ---------------------------------------------------------------------------
// Invalidation (per-user generation)
// ---------------------------------------------------------------------------

// fullyLoadedRepo returns a fake preloaded so the first List/Count calls
// populate the caches, and a CachedRepository over it.
func fullyLoadedRepo(now time.Time) (*fakeRepo, *CachedRepository) {
	fake := &fakeRepo{
		listResult:   []*residentlocationdomain.ResidentLocation{newTestLocation(1, "user-1", now)},
		countResult:  1,
		createResult: newTestLocation(2, "user-1", now),
		updateResult: newTestLocation(1, "user-1", now),
		deleteResult: 1,
	}
	return fake, NewCachedRepository(fake)
}

func TestCachedRepository_InvalidatedByCreate(t *testing.T) {
	t.Parallel()

	now := time.Date(2024, 6, 15, 10, 30, 0, 0, time.UTC)
	fake, repo := fullyLoadedRepo(now)
	ctx := context.Background()

	_, err := repo.List(ctx, 20, 0, "user-1")
	assert.NoError(t, err)
	_, err = repo.Count(ctx, "user-1")
	assert.NoError(t, err)
	assert.Equal(t, 1, fake.listCalls)
	assert.Equal(t, 1, fake.countCalls)

	created, err := repo.Create(ctx, residentlocationdomain.NewResidentLocationInput("Belarus", "Minsk", "220000", "Nezavisimosti", "10", "5"), "user-1")
	assert.NoError(t, err)
	assertCachedResidentLocationEqual(t, fake.createResult, created)

	_, err = repo.List(ctx, 20, 0, "user-1")
	assert.NoError(t, err)
	_, err = repo.Count(ctx, "user-1")
	assert.NoError(t, err)
	assert.Equal(t, 2, fake.listCalls, "Create must invalidate cached lists")
	assert.Equal(t, 2, fake.countCalls, "Create must invalidate cached counts")
}

func TestCachedRepository_InvalidatedByUpdate(t *testing.T) {
	t.Parallel()

	now := time.Date(2024, 6, 15, 10, 30, 0, 0, time.UTC)
	fake, repo := fullyLoadedRepo(now)
	ctx := context.Background()

	_, err := repo.List(ctx, 20, 0, "user-1")
	assert.NoError(t, err)
	assert.Equal(t, 1, fake.listCalls)

	updated, err := repo.Update(ctx, 1, residentlocationdomain.NewResidentLocationInput("Belarus", "Minsk", "220000", "Nezavisimosti", "10", "5"), "user-1")
	assert.NoError(t, err)
	assertCachedResidentLocationEqual(t, fake.updateResult, updated)

	_, err = repo.List(ctx, 20, 0, "user-1")
	assert.NoError(t, err)
	assert.Equal(t, 2, fake.listCalls, "Update must invalidate cached lists")
}

func TestCachedRepository_InvalidatedByDelete(t *testing.T) {
	t.Parallel()

	now := time.Date(2024, 6, 15, 10, 30, 0, 0, time.UTC)
	fake, repo := fullyLoadedRepo(now)
	ctx := context.Background()

	_, err := repo.List(ctx, 20, 0, "user-1")
	assert.NoError(t, err)
	assert.Equal(t, 1, fake.listCalls)

	deleted, err := repo.Delete(ctx, 1, "user-1")
	assert.NoError(t, err)
	assert.Equal(t, int64(1), deleted)

	_, err = repo.List(ctx, 20, 0, "user-1")
	assert.NoError(t, err)
	assert.Equal(t, 2, fake.listCalls, "Delete must invalidate cached lists")
}

// TestCachedRepository_PerUserIsolation verifies a write by one user does NOT
// invalidate another user's cached entries.
func TestCachedRepository_PerUserIsolation(t *testing.T) {
	t.Parallel()

	now := time.Date(2024, 6, 15, 10, 30, 0, 0, time.UTC)
	fake := &fakeRepo{
		listResult:   []*residentlocationdomain.ResidentLocation{newTestLocation(1, "user-1", now)},
		createResult: newTestLocation(2, "user-2", now),
	}
	repo := NewCachedRepository(fake)
	ctx := context.Background()

	_, err := repo.List(ctx, 20, 0, "user-1")
	assert.NoError(t, err)
	_, err = repo.List(ctx, 20, 0, "user-2")
	assert.NoError(t, err)
	assert.Equal(t, 2, fake.listCalls)

	_, err = repo.Create(ctx, residentlocationdomain.NewResidentLocationInput("Belarus", "Minsk", "220000", "Nezavisimosti", "10", "5"), "user-2")
	assert.NoError(t, err)

	_, err = repo.List(ctx, 20, 0, "user-2")
	assert.NoError(t, err)
	assert.Equal(t, 3, fake.listCalls, "writer's own cache must be invalidated")

	_, err = repo.List(ctx, 20, 0, "user-1")
	assert.NoError(t, err)
	assert.Equal(t, 3, fake.listCalls, "other user's cache must NOT be invalidated")
}

func TestCachedRepository_FailedWriteDoesNotInvalidate(t *testing.T) {
	t.Parallel()

	now := time.Date(2024, 6, 15, 10, 30, 0, 0, time.UTC)
	fake := &fakeRepo{
		listResult: []*residentlocationdomain.ResidentLocation{newTestLocation(1, "user-1", now)},
		createErr:  errors.New("insert failed"),
	}
	repo := NewCachedRepository(fake)
	ctx := context.Background()

	_, err := repo.List(ctx, 20, 0, "user-1")
	assert.NoError(t, err)
	assert.Equal(t, 1, fake.listCalls)

	_, err = repo.Create(ctx, residentlocationdomain.NewResidentLocationInput("Belarus", "Minsk", "220000", "Nezavisimosti", "10", "5"), "user-1")
	assert.Error(t, err)

	_, err = repo.List(ctx, 20, 0, "user-1")
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
	items := []*residentlocationdomain.ResidentLocation{newTestLocation(1, "user-1", now)}
	fake := &fakeRepo{listResult: items, countResult: 1}
	repo := &CachedRepository{repo: fake} // caches intentionally nil
	ctx := context.Background()

	got, err := repo.List(ctx, 20, 0, "user-1")
	assert.NoError(t, err)
	assertCachedResidentLocationSliceEqual(t, items, got)
	assert.Equal(t, 1, fake.listCalls)

	count, err := repo.Count(ctx, "user-1")
	assert.NoError(t, err)
	assert.Equal(t, 1, count)
	assert.Equal(t, 1, fake.countCalls)
}
