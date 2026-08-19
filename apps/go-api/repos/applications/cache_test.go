package applications

import (
	"context"
	"errors"
	"testing"
	"time"

	applicationsdomain "flatty-budget/go-api/domains/applications"
	"github.com/stretchr/testify/assert"
)

// ---------------------------------------------------------------------------
// Fake repository
// ---------------------------------------------------------------------------

// fakeRepo implements applicationsdomain.Repository in memory, counting calls
// so tests can assert cache hits bypass the wrapped repository.
type fakeRepo struct {
	listByEnvCalls int
	listAllCalls   int
	createCalls    int
	updateCalls    int
	deleteCalls    int

	listByEnvResult []*applicationsdomain.Application
	listByEnvErr    error
	listAllResult   []*applicationsdomain.Application
	listAllErr      error

	createResult *applicationsdomain.Application
	createErr    error
	updateResult *applicationsdomain.Application
	updateErr    error
	deleteResult int64
	deleteErr    error
}

func (f *fakeRepo) ListByEnv(ctx context.Context, env string) ([]*applicationsdomain.Application, error) {
	f.listByEnvCalls++
	return f.listByEnvResult, f.listByEnvErr
}

func (f *fakeRepo) ListAll(ctx context.Context) ([]*applicationsdomain.Application, error) {
	f.listAllCalls++
	return f.listAllResult, f.listAllErr
}

func (f *fakeRepo) Create(ctx context.Context, input *applicationsdomain.ApplicationInput) (*applicationsdomain.Application, error) {
	f.createCalls++
	return f.createResult, f.createErr
}

func (f *fakeRepo) Update(ctx context.Context, id int64, input *applicationsdomain.ApplicationInput) (*applicationsdomain.Application, error) {
	f.updateCalls++
	return f.updateResult, f.updateErr
}

func (f *fakeRepo) Delete(ctx context.Context, id int64) (int64, error) {
	f.deleteCalls++
	return f.deleteResult, f.deleteErr
}

// ---------------------------------------------------------------------------
// Assertion helpers
// ---------------------------------------------------------------------------

func assertApplicationEqual(t *testing.T, want, got *applicationsdomain.Application) {
	t.Helper()
	assert.Equal(t, want.ID(), got.ID())
	assert.Equal(t, want.Name(), got.Name())
	assert.Equal(t, want.Env(), got.Env())
	assert.Equal(t, want.BundleJS(), got.BundleJS())
	assert.Equal(t, want.StyleURL(), got.StyleURL())
	assert.Equal(t, want.RemoteOrigin(), got.RemoteOrigin())
	assert.Equal(t, want.ProxyBasePath(), got.ProxyBasePath())
	assert.Equal(t, want.BasePath(), got.BasePath())
	assert.True(t, want.CreatedAt().Equal(got.CreatedAt()), "CreatedAt mismatch")
	assert.True(t, want.UpdatedAt().Equal(got.UpdatedAt()), "UpdatedAt mismatch")
}

func assertApplicationSliceEqual(t *testing.T, want, got []*applicationsdomain.Application) {
	t.Helper()
	assert.Equal(t, len(want), len(got))
	for i := range want {
		assertApplicationEqual(t, want[i], got[i])
	}
}

func newTestApp(id int64, name, env string, now time.Time) *applicationsdomain.Application {
	return applicationsdomain.NewApplication(
		id, name, env, "http://cdn/"+name+".js", "http://cdn/"+name+".css",
		"http://origin/"+name, "/proxy/"+name, "/base/"+name,
		now, now,
	)
}

// ---------------------------------------------------------------------------
// ListByEnv
// ---------------------------------------------------------------------------

// TestCachedRepository_ListByEnv_CacheHit covers both the miss (first call
// populates) and the hit (second call is served from cache — wrapped repo
// untouched).
func TestCachedRepository_ListByEnv_CacheHit(t *testing.T) {
	t.Parallel()

	now := time.Date(2024, 6, 15, 10, 30, 0, 0, time.UTC)
	apps := []*applicationsdomain.Application{newTestApp(1, "launcher", "prod", now)}
	fake := &fakeRepo{listByEnvResult: apps}
	repo := NewCachedRepository(fake)
	ctx := context.Background()

	first, err := repo.ListByEnv(ctx, "prod")
	assert.NoError(t, err)
	assertApplicationSliceEqual(t, apps, first)
	assert.Equal(t, 1, fake.listByEnvCalls)

	second, err := repo.ListByEnv(ctx, "prod")
	assert.NoError(t, err)
	assertApplicationSliceEqual(t, apps, second)
	assert.Equal(t, 1, fake.listByEnvCalls, "second ListByEnv with same args must be served from cache")
}

func TestCachedRepository_ListByEnv_KeyIsolation(t *testing.T) {
	t.Parallel()

	now := time.Date(2024, 6, 15, 10, 30, 0, 0, time.UTC)
	fake := &fakeRepo{listByEnvResult: []*applicationsdomain.Application{newTestApp(1, "launcher", "prod", now)}}
	repo := NewCachedRepository(fake)
	ctx := context.Background()

	_, err := repo.ListByEnv(ctx, "prod")
	assert.NoError(t, err)
	_, err = repo.ListByEnv(ctx, "staging")
	assert.NoError(t, err)
	_, err = repo.ListByEnv(ctx, "prod") // repeat: must be served from cache
	assert.NoError(t, err)

	assert.Equal(t, 2, fake.listByEnvCalls, "each distinct env must have its own cache entry")
}

func TestCachedRepository_ListByEnv_DefensiveCopy(t *testing.T) {
	t.Parallel()

	now := time.Date(2024, 6, 15, 10, 30, 0, 0, time.UTC)
	apps := []*applicationsdomain.Application{newTestApp(1, "launcher", "prod", now)}
	fake := &fakeRepo{listByEnvResult: apps}
	repo := NewCachedRepository(fake)
	ctx := context.Background()

	first, err := repo.ListByEnv(ctx, "prod")
	assert.NoError(t, err)
	first = append(first, newTestApp(99, "Injected", "prod", now))

	second, err := repo.ListByEnv(ctx, "prod")
	assert.NoError(t, err)
	assert.Equal(t, 1, len(second), "mutating a returned slice must not corrupt the cached entry")
}

func TestCachedRepository_ListByEnv_DBErrorNotCached(t *testing.T) {
	t.Parallel()

	now := time.Date(2024, 6, 15, 10, 30, 0, 0, time.UTC)
	apps := []*applicationsdomain.Application{newTestApp(1, "launcher", "prod", now)}
	fake := &fakeRepo{listByEnvErr: errors.New("db down")}
	repo := NewCachedRepository(fake)
	ctx := context.Background()

	_, err := repo.ListByEnv(ctx, "prod")
	assert.Error(t, err)

	fake.listByEnvErr = nil
	fake.listByEnvResult = apps

	got, err := repo.ListByEnv(ctx, "prod")
	assert.NoError(t, err)
	assertApplicationSliceEqual(t, apps, got)
	assert.Equal(t, 2, fake.listByEnvCalls, "DB errors must never be cached")
}

func TestCachedRepository_ListByEnv_EmptyResultIsCached(t *testing.T) {
	t.Parallel()

	fake := &fakeRepo{listByEnvResult: nil}
	repo := NewCachedRepository(fake)
	ctx := context.Background()

	first, err := repo.ListByEnv(ctx, "prod")
	assert.NoError(t, err)
	assert.Nil(t, first)
	assert.Equal(t, 1, fake.listByEnvCalls)

	second, err := repo.ListByEnv(ctx, "prod")
	assert.NoError(t, err)
	assert.Nil(t, second)
	assert.Equal(t, 1, fake.listByEnvCalls, "empty results are valid and must be cached")
}

func TestCachedRepository_ListByEnv_TTLExpiry(t *testing.T) {
	t.Parallel()

	now := time.Date(2024, 6, 15, 10, 30, 0, 0, time.UTC)
	apps := []*applicationsdomain.Application{newTestApp(1, "launcher", "prod", now)}
	fake := &fakeRepo{listByEnvResult: apps}
	repo := NewCachedRepositoryWithTTL(fake, 20*time.Millisecond)
	ctx := context.Background()

	_, err := repo.ListByEnv(ctx, "prod")
	assert.NoError(t, err)
	_, err = repo.ListByEnv(ctx, "prod")
	assert.NoError(t, err)
	assert.Equal(t, 1, fake.listByEnvCalls, "calls within TTL are served from cache")

	time.Sleep(60 * time.Millisecond)

	_, err = repo.ListByEnv(ctx, "prod")
	assert.NoError(t, err)
	assert.Equal(t, 2, fake.listByEnvCalls, "entry must expire and re-fetch after TTL")
}

// ---------------------------------------------------------------------------
// ListAll
// ---------------------------------------------------------------------------

func TestCachedRepository_ListAll_CacheHit(t *testing.T) {
	t.Parallel()

	now := time.Date(2024, 6, 15, 10, 30, 0, 0, time.UTC)
	apps := []*applicationsdomain.Application{newTestApp(1, "launcher", "prod", now)}
	fake := &fakeRepo{listAllResult: apps}
	repo := NewCachedRepository(fake)
	ctx := context.Background()

	first, err := repo.ListAll(ctx)
	assert.NoError(t, err)
	assertApplicationSliceEqual(t, apps, first)
	assert.Equal(t, 1, fake.listAllCalls)

	second, err := repo.ListAll(ctx)
	assert.NoError(t, err)
	assertApplicationSliceEqual(t, apps, second)
	assert.Equal(t, 1, fake.listAllCalls, "second ListAll must be served from cache")
}

func TestCachedRepository_ListAll_DBErrorNotCached(t *testing.T) {
	t.Parallel()

	now := time.Date(2024, 6, 15, 10, 30, 0, 0, time.UTC)
	apps := []*applicationsdomain.Application{newTestApp(1, "launcher", "prod", now)}
	fake := &fakeRepo{listAllErr: errors.New("db down")}
	repo := NewCachedRepository(fake)
	ctx := context.Background()

	_, err := repo.ListAll(ctx)
	assert.Error(t, err)

	fake.listAllErr = nil
	fake.listAllResult = apps

	got, err := repo.ListAll(ctx)
	assert.NoError(t, err)
	assertApplicationSliceEqual(t, apps, got)
	assert.Equal(t, 2, fake.listAllCalls, "DB errors must never be cached")
}

// ---------------------------------------------------------------------------
// Invalidation
// ---------------------------------------------------------------------------

// invalidatingRepo returns a fake preloaded so the first ListByEnv/ListAll
// calls populate the caches, and a CachedRepository over it.
func invalidatingRepo(now time.Time) (*fakeRepo, *CachedRepository) {
	fake := &fakeRepo{
		listByEnvResult: []*applicationsdomain.Application{newTestApp(1, "launcher", "prod", now)},
		listAllResult:   []*applicationsdomain.Application{newTestApp(1, "launcher", "prod", now)},
		createResult:    newTestApp(2, "settings", "prod", now),
		updateResult:    newTestApp(1, "launcher-updated", "prod", now),
		deleteResult:    1,
	}
	return fake, NewCachedRepository(fake)
}

func TestCachedRepository_InvalidatedByCreate(t *testing.T) {
	t.Parallel()

	now := time.Date(2024, 6, 15, 10, 30, 0, 0, time.UTC)
	fake, repo := invalidatingRepo(now)
	ctx := context.Background()

	_, err := repo.ListByEnv(ctx, "prod")
	assert.NoError(t, err)
	_, err = repo.ListAll(ctx)
	assert.NoError(t, err)
	assert.Equal(t, 1, fake.listByEnvCalls)
	assert.Equal(t, 1, fake.listAllCalls)

	created, err := repo.Create(ctx, applicationsdomain.NewApplicationInput(
		"settings", "prod", "http://cdn/settings.js", "http://cdn/settings.css",
		"http://origin/settings", "/proxy/settings", "/base/settings",
	))
	assert.NoError(t, err)
	assertApplicationEqual(t, fake.createResult, created)

	_, err = repo.ListByEnv(ctx, "prod")
	assert.NoError(t, err)
	_, err = repo.ListAll(ctx)
	assert.NoError(t, err)
	assert.Equal(t, 2, fake.listByEnvCalls, "Create must invalidate cached lists")
	assert.Equal(t, 2, fake.listAllCalls, "Create must invalidate cached lists")
}

func TestCachedRepository_InvalidatedByUpdate(t *testing.T) {
	t.Parallel()

	now := time.Date(2024, 6, 15, 10, 30, 0, 0, time.UTC)
	fake, repo := invalidatingRepo(now)
	ctx := context.Background()

	_, err := repo.ListByEnv(ctx, "prod")
	assert.NoError(t, err)
	_, err = repo.ListAll(ctx)
	assert.NoError(t, err)
	assert.Equal(t, 1, fake.listByEnvCalls)
	assert.Equal(t, 1, fake.listAllCalls)

	updated, err := repo.Update(ctx, 1, applicationsdomain.NewApplicationInput(
		"launcher-updated", "prod", "http://cdn/launcher.js", "http://cdn/launcher.css",
		"http://origin/launcher", "/proxy/launcher", "/base/launcher",
	))
	assert.NoError(t, err)
	assertApplicationEqual(t, fake.updateResult, updated)

	_, err = repo.ListByEnv(ctx, "prod")
	assert.NoError(t, err)
	_, err = repo.ListAll(ctx)
	assert.NoError(t, err)
	assert.Equal(t, 2, fake.listByEnvCalls, "Update must invalidate cached lists")
	assert.Equal(t, 2, fake.listAllCalls, "Update must invalidate cached lists")
}

func TestCachedRepository_InvalidatedByDelete(t *testing.T) {
	t.Parallel()

	now := time.Date(2024, 6, 15, 10, 30, 0, 0, time.UTC)
	fake, repo := invalidatingRepo(now)
	ctx := context.Background()

	_, err := repo.ListByEnv(ctx, "prod")
	assert.NoError(t, err)
	_, err = repo.ListAll(ctx)
	assert.NoError(t, err)
	assert.Equal(t, 1, fake.listByEnvCalls)
	assert.Equal(t, 1, fake.listAllCalls)

	deleted, err := repo.Delete(ctx, 1)
	assert.NoError(t, err)
	assert.Equal(t, int64(1), deleted)

	_, err = repo.ListByEnv(ctx, "prod")
	assert.NoError(t, err)
	_, err = repo.ListAll(ctx)
	assert.NoError(t, err)
	assert.Equal(t, 2, fake.listByEnvCalls, "Delete must invalidate cached lists")
	assert.Equal(t, 2, fake.listAllCalls, "Delete must invalidate cached lists")
}

func TestCachedRepository_FailedWriteDoesNotInvalidate(t *testing.T) {
	t.Parallel()

	now := time.Date(2024, 6, 15, 10, 30, 0, 0, time.UTC)
	fake := &fakeRepo{
		listByEnvResult: []*applicationsdomain.Application{newTestApp(1, "launcher", "prod", now)},
		listAllResult:   []*applicationsdomain.Application{newTestApp(1, "launcher", "prod", now)},
		createErr:       errors.New("insert failed"),
	}
	repo := NewCachedRepository(fake)
	ctx := context.Background()

	_, err := repo.ListByEnv(ctx, "prod")
	assert.NoError(t, err)
	_, err = repo.ListAll(ctx)
	assert.NoError(t, err)
	assert.Equal(t, 1, fake.listByEnvCalls)
	assert.Equal(t, 1, fake.listAllCalls)

	_, err = repo.Create(ctx, applicationsdomain.NewApplicationInput(
		"settings", "prod", "http://cdn/settings.js", "http://cdn/settings.css",
		"http://origin/settings", "/proxy/settings", "/base/settings",
	))
	assert.Error(t, err)

	_, err = repo.ListByEnv(ctx, "prod")
	assert.NoError(t, err)
	_, err = repo.ListAll(ctx)
	assert.NoError(t, err)
	assert.Equal(t, 1, fake.listByEnvCalls, "failed Create must not bump the generation counter")
	assert.Equal(t, 1, fake.listAllCalls, "failed Create must not bump the generation counter")
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
	apps := []*applicationsdomain.Application{newTestApp(1, "launcher", "prod", now)}
	fake := &fakeRepo{listByEnvResult: apps, listAllResult: apps}
	repo := &CachedRepository{repo: fake} // caches intentionally nil
	ctx := context.Background()

	got, err := repo.ListByEnv(ctx, "prod")
	assert.NoError(t, err)
	assertApplicationSliceEqual(t, apps, got)
	assert.Equal(t, 1, fake.listByEnvCalls)

	all, err := repo.ListAll(ctx)
	assert.NoError(t, err)
	assertApplicationSliceEqual(t, apps, all)
	assert.Equal(t, 1, fake.listAllCalls)
}
