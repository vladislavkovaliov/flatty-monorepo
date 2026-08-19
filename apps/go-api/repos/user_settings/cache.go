package user_settings

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/hashicorp/golang-lru/v2/expirable"

	user_settings_domain "flatty-budget/go-api/domains/user_settings"
)

// Cache defaults for the CachedRepository decorator.
const (
	defaultCacheSize = 1024
	defaultCacheTTL  = time.Hour
)

// CachedRepository decorates a user_settings_domain.UserSettingsRepository with
// an in-memory TTL cache for GetByUserID results. Upsert bumps a per-user
// generation counter embedded in the cache key, which invalidates that user's
// cached entry on the next read. Thread-safe via the underlying LRU and a
// mutex-guarded generation map.
//
// A nil result (user has no settings row yet) is a valid, cacheable outcome:
// the LRU stores a nil value and reports a hit, so the wrapped repository is
// not re-queried until the entry expires or Upsert bumps the generation.
//
// Fail-open: if the cache is nil (or an entry is absent) the decorator simply
// delegates to the wrapped repository — a cache problem can never break the
// request path. DB errors are never cached.
type CachedRepository struct {
	repo          user_settings_domain.UserSettingsRepository
	settingsCache *expirable.LRU[string, *user_settings_domain.UserSettings]
	mu            sync.Mutex
	gen           map[string]int64
}

// NewCachedRepository wraps repo with a default in-memory LRU cache (1024
// entries, 1 hour TTL).
func NewCachedRepository(repo user_settings_domain.UserSettingsRepository) *CachedRepository {
	return NewCachedRepositoryWithTTL(repo, defaultCacheTTL)
}

// NewCachedRepositoryWithTTL wraps repo with an in-memory LRU cache using the
// given TTL. Exposed for tests (tiny TTLs) — production code uses
// NewCachedRepository.
func NewCachedRepositoryWithTTL(repo user_settings_domain.UserSettingsRepository, ttl time.Duration) *CachedRepository {
	return &CachedRepository{
		repo:          repo,
		settingsCache: expirable.NewLRU[string, *user_settings_domain.UserSettings](defaultCacheSize, nil, ttl),
	}
}

// GetByUserID returns a cached entry when available; on a miss it fetches from
// the wrapped repository and stores the result. A nil result (no settings row)
// is cached too. DB errors are never cached.
func (r *CachedRepository) GetByUserID(ctx context.Context, userID string) (*user_settings_domain.UserSettings, error) {
	if r.settingsCache == nil {
		return r.repo.GetByUserID(ctx, userID)
	}

	key := settingsCacheKey(r.generation(userID), userID)
	if cached, ok := r.settingsCache.Get(key); ok {
		return cached, nil
	}

	settings, err := r.repo.GetByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	r.settingsCache.Add(key, settings)
	return settings, nil
}

// Upsert writes through to the wrapped repository, then bumps the user's
// generation counter on success (invalidating that user's cached entry).
func (r *CachedRepository) Upsert(ctx context.Context, userID string, input *user_settings_domain.UserSettingsInput) (*user_settings_domain.UserSettings, error) {
	settings, err := r.repo.Upsert(ctx, userID, input)
	if err != nil {
		return nil, err
	}
	r.bump(userID)
	return settings, nil
}

// generation returns the current generation counter for user. Safe to call on
// a nil map (returns 0); entries are created lazily by bump.
func (r *CachedRepository) generation(userID string) int64 {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.gen[userID]
}

// bump increments the user's generation counter, creating the map entry (and
// the map itself) on first use.
func (r *CachedRepository) bump(userID string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.gen == nil {
		r.gen = make(map[string]int64)
	}
	r.gen[userID]++
}

// settingsCacheKey embeds the user's generation counter so any Upsert by that
// user invalidates the cached entry without a prefix scan.
func settingsCacheKey(gen int64, userID string) string {
	return fmt.Sprintf("user-settings:%s:%d", userID, gen)
}

// Compile-time check: CachedRepository satisfies
// user_settings_domain.UserSettingsRepository.
var _ user_settings_domain.UserSettingsRepository = (*CachedRepository)(nil)
