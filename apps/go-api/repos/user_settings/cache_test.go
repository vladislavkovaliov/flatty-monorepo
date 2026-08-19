package user_settings

import (
	"context"
	"errors"
	"testing"
	"time"

	user_settings_domain "flatty-budget/go-api/domains/user_settings"
	"github.com/stretchr/testify/assert"
)

// ---------------------------------------------------------------------------
// Fake repository
// ---------------------------------------------------------------------------

// fakeRepo implements user_settings_domain.UserSettingsRepository in memory,
// counting calls so tests can assert cache hits bypass the wrapped repository.
type fakeRepo struct {
	getByUserIDCalls int
	upsertCalls      int

	getByUserIDResult *user_settings_domain.UserSettings
	getByUserIDErr    error
	upsertResult      *user_settings_domain.UserSettings
	upsertErr         error
}

func (f *fakeRepo) GetByUserID(ctx context.Context, userID string) (*user_settings_domain.UserSettings, error) {
	f.getByUserIDCalls++
	return f.getByUserIDResult, f.getByUserIDErr
}

func (f *fakeRepo) Upsert(ctx context.Context, userID string, input *user_settings_domain.UserSettingsInput) (*user_settings_domain.UserSettings, error) {
	f.upsertCalls++
	return f.upsertResult, f.upsertErr
}

// ---------------------------------------------------------------------------
// Assertion helpers
// ---------------------------------------------------------------------------

func assertCachedUserSettingsEqual(t *testing.T, want, got *user_settings_domain.UserSettings) {
	t.Helper()
	assert.Equal(t, want.UserID(), got.UserID())
	assert.Equal(t, want.Language(), got.Language())
	assert.Equal(t, want.Theme(), got.Theme())
	assert.Equal(t, want.Timezone(), got.Timezone())
	assert.Equal(t, want.DateFormat(), got.DateFormat())
	assert.True(t, want.CreatedAt().Equal(got.CreatedAt()), "CreatedAt mismatch")
	assert.True(t, want.UpdatedAt().Equal(got.UpdatedAt()), "UpdatedAt mismatch")
}

func newTestSettings(userID string, now time.Time) *user_settings_domain.UserSettings {
	return user_settings_domain.NewUserSettings(userID, "en", "light", "UTC", "DD/MM/YYYY", now, now)
}

// ---------------------------------------------------------------------------
// GetByUserID
// ---------------------------------------------------------------------------

// TestCachedRepository_GetByUserID_CacheHit covers both the miss (first call
// populates) and the hit (second call is served from cache — wrapped repo
// untouched).
func TestCachedRepository_GetByUserID_CacheHit(t *testing.T) {
	t.Parallel()

	now := time.Date(2024, 6, 15, 10, 30, 0, 0, time.UTC)
	settings := newTestSettings("user-1", now)
	fake := &fakeRepo{getByUserIDResult: settings}
	repo := NewCachedRepository(fake)
	ctx := context.Background()

	first, err := repo.GetByUserID(ctx, "user-1")
	assert.NoError(t, err)
	assertCachedUserSettingsEqual(t, settings, first)
	assert.Equal(t, 1, fake.getByUserIDCalls)

	second, err := repo.GetByUserID(ctx, "user-1")
	assert.NoError(t, err)
	assertCachedUserSettingsEqual(t, settings, second)
	assert.Equal(t, 1, fake.getByUserIDCalls, "second GetByUserID with same args must be served from cache")
}

func TestCachedRepository_GetByUserID_KeyIsolation(t *testing.T) {
	t.Parallel()

	now := time.Date(2024, 6, 15, 10, 30, 0, 0, time.UTC)
	fake := &fakeRepo{getByUserIDResult: newTestSettings("user-1", now)}
	repo := NewCachedRepository(fake)
	ctx := context.Background()

	_, err := repo.GetByUserID(ctx, "user-1")
	assert.NoError(t, err)
	_, err = repo.GetByUserID(ctx, "user-2")
	assert.NoError(t, err)
	_, err = repo.GetByUserID(ctx, "user-1") // repeat: cache hit
	assert.NoError(t, err)

	assert.Equal(t, 2, fake.getByUserIDCalls, "each distinct userID must have its own cache entry")
}

// TestCachedRepository_GetByUserID_NilResultIsCached verifies that a nil
// result (no settings row yet) is a valid cacheable outcome — the second call
// must NOT re-query the wrapped repository.
func TestCachedRepository_GetByUserID_NilResultIsCached(t *testing.T) {
	t.Parallel()

	fake := &fakeRepo{getByUserIDResult: nil}
	repo := NewCachedRepository(fake)
	ctx := context.Background()

	first, err := repo.GetByUserID(ctx, "user-1")
	assert.NoError(t, err)
	assert.Nil(t, first)
	assert.Equal(t, 1, fake.getByUserIDCalls)

	second, err := repo.GetByUserID(ctx, "user-1")
	assert.NoError(t, err)
	assert.Nil(t, second)
	assert.Equal(t, 1, fake.getByUserIDCalls, "nil results are valid and must be cached")
}

func TestCachedRepository_GetByUserID_DBErrorNotCached(t *testing.T) {
	t.Parallel()

	now := time.Date(2024, 6, 15, 10, 30, 0, 0, time.UTC)
	settings := newTestSettings("user-1", now)
	fake := &fakeRepo{getByUserIDErr: errors.New("db down")}
	repo := NewCachedRepository(fake)
	ctx := context.Background()

	_, err := repo.GetByUserID(ctx, "user-1")
	assert.Error(t, err)

	fake.getByUserIDErr = nil
	fake.getByUserIDResult = settings

	got, err := repo.GetByUserID(ctx, "user-1")
	assert.NoError(t, err)
	assertCachedUserSettingsEqual(t, settings, got)
	assert.Equal(t, 2, fake.getByUserIDCalls, "DB errors must never be cached")
}

func TestCachedRepository_GetByUserID_TTLExpiry(t *testing.T) {
	t.Parallel()

	now := time.Date(2024, 6, 15, 10, 30, 0, 0, time.UTC)
	settings := newTestSettings("user-1", now)
	fake := &fakeRepo{getByUserIDResult: settings}
	repo := NewCachedRepositoryWithTTL(fake, 20*time.Millisecond)
	ctx := context.Background()

	_, err := repo.GetByUserID(ctx, "user-1")
	assert.NoError(t, err)
	_, err = repo.GetByUserID(ctx, "user-1")
	assert.NoError(t, err)
	assert.Equal(t, 1, fake.getByUserIDCalls, "calls within TTL are served from cache")

	time.Sleep(60 * time.Millisecond)

	_, err = repo.GetByUserID(ctx, "user-1")
	assert.NoError(t, err)
	assert.Equal(t, 2, fake.getByUserIDCalls, "entry must expire and re-fetch after TTL")
}

// ---------------------------------------------------------------------------
// Invalidation (per-user generation)
// ---------------------------------------------------------------------------

func TestCachedRepository_InvalidatedByUpsert(t *testing.T) {
	t.Parallel()

	now := time.Date(2024, 6, 15, 10, 30, 0, 0, time.UTC)
	fake := &fakeRepo{
		getByUserIDResult: newTestSettings("user-1", now),
		upsertResult:      newTestSettings("user-1", now),
	}
	repo := NewCachedRepository(fake)
	ctx := context.Background()

	_, err := repo.GetByUserID(ctx, "user-1")
	assert.NoError(t, err)
	assert.Equal(t, 1, fake.getByUserIDCalls)

	upserted, err := repo.Upsert(ctx, "user-1", user_settings_domain.NewUserSettingsInput("fr", "dark", "Europe/Paris", "DD/MM/YYYY"))
	assert.NoError(t, err)
	assertCachedUserSettingsEqual(t, fake.upsertResult, upserted)

	_, err = repo.GetByUserID(ctx, "user-1")
	assert.NoError(t, err)
	assert.Equal(t, 2, fake.getByUserIDCalls, "Upsert must invalidate the cached entry")
}

// TestCachedRepository_PerUserIsolation verifies an Upsert by one user does NOT
// invalidate another user's cached entry.
func TestCachedRepository_PerUserIsolation(t *testing.T) {
	t.Parallel()

	now := time.Date(2024, 6, 15, 10, 30, 0, 0, time.UTC)
	fake := &fakeRepo{
		getByUserIDResult: newTestSettings("user-1", now),
		upsertResult:      newTestSettings("user-2", now),
	}
	repo := NewCachedRepository(fake)
	ctx := context.Background()

	_, err := repo.GetByUserID(ctx, "user-1")
	assert.NoError(t, err)
	_, err = repo.GetByUserID(ctx, "user-2")
	assert.NoError(t, err)
	assert.Equal(t, 2, fake.getByUserIDCalls)

	_, err = repo.Upsert(ctx, "user-2", user_settings_domain.NewUserSettingsInput("fr", "dark", "Europe/Paris", "DD/MM/YYYY"))
	assert.NoError(t, err)

	_, err = repo.GetByUserID(ctx, "user-2")
	assert.NoError(t, err)
	assert.Equal(t, 3, fake.getByUserIDCalls, "writer's own cache must be invalidated")

	_, err = repo.GetByUserID(ctx, "user-1")
	assert.NoError(t, err)
	assert.Equal(t, 3, fake.getByUserIDCalls, "other user's cache must NOT be invalidated")
}

func TestCachedRepository_FailedUpsertDoesNotInvalidate(t *testing.T) {
	t.Parallel()

	now := time.Date(2024, 6, 15, 10, 30, 0, 0, time.UTC)
	fake := &fakeRepo{
		getByUserIDResult: newTestSettings("user-1", now),
		upsertErr:         errors.New("upsert failed"),
	}
	repo := NewCachedRepository(fake)
	ctx := context.Background()

	_, err := repo.GetByUserID(ctx, "user-1")
	assert.NoError(t, err)
	assert.Equal(t, 1, fake.getByUserIDCalls)

	_, err = repo.Upsert(ctx, "user-1", user_settings_domain.NewUserSettingsInput("fr", "dark", "Europe/Paris", "DD/MM/YYYY"))
	assert.Error(t, err)

	_, err = repo.GetByUserID(ctx, "user-1")
	assert.NoError(t, err)
	assert.Equal(t, 1, fake.getByUserIDCalls, "failed Upsert must not bump the generation counter")
}

// ---------------------------------------------------------------------------
// Fail-open
// ---------------------------------------------------------------------------

// TestCachedRepository_FailOpen_NilCache verifies that a CachedRepository with
// a nil cache (the only realistic failure mode for an in-memory LRU) bypasses
// caching entirely and delegates to the wrapped repository.
func TestCachedRepository_FailOpen_NilCache(t *testing.T) {
	t.Parallel()

	now := time.Date(2024, 6, 15, 10, 30, 0, 0, time.UTC)
	settings := newTestSettings("user-1", now)
	fake := &fakeRepo{getByUserIDResult: settings}
	repo := &CachedRepository{repo: fake} // cache intentionally nil
	ctx := context.Background()

	got, err := repo.GetByUserID(ctx, "user-1")
	assert.NoError(t, err)
	assertCachedUserSettingsEqual(t, settings, got)
	assert.Equal(t, 1, fake.getByUserIDCalls)
}
