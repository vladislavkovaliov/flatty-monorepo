package residentlocation

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/hashicorp/golang-lru/v2/expirable"

	residentlocationdomain "flatty-budget/go-api/domains/resident_location"
)

// Cache defaults for the CachedRepository decorator.
const (
	defaultCacheSize = 1024
	defaultCacheTTL  = time.Hour
)

// CachedRepository decorates a residentlocationdomain.Repository with an
// in-memory TTL cache for Count/List results. Write operations (Create/Update/
// Delete) bump a per-user generation counter embedded in every cache key, which
// invalidates only that user's cached entries on the next read. Thread-safe via
// the underlying LRUs and a mutex-guarded generation map.
//
// Fail-open: if the caches are nil (or an entry is absent) the decorator simply
// delegates to the wrapped repository — a cache problem can never break the
// request path. DB errors are never cached.
type CachedRepository struct {
	repo       residentlocationdomain.Repository
	listCache  *expirable.LRU[string, []*residentlocationdomain.ResidentLocation]
	countCache *expirable.LRU[string, int]
	mu         sync.Mutex
	gen        map[string]int64
}

// NewCachedRepository wraps repo with a default in-memory LRU cache (1024
// entries, 1 hour TTL).
func NewCachedRepository(repo residentlocationdomain.Repository) *CachedRepository {
	return NewCachedRepositoryWithTTL(repo, defaultCacheTTL)
}

// NewCachedRepositoryWithTTL wraps repo with an in-memory LRU cache using the
// given TTL. Exposed for tests (tiny TTLs) — production code uses
// NewCachedRepository.
func NewCachedRepositoryWithTTL(repo residentlocationdomain.Repository, ttl time.Duration) *CachedRepository {
	return &CachedRepository{
		repo:       repo,
		listCache:  expirable.NewLRU[string, []*residentlocationdomain.ResidentLocation](defaultCacheSize, nil, ttl),
		countCache: expirable.NewLRU[string, int](defaultCacheSize, nil, ttl),
	}
}

// List returns a cached page when available; on a miss it fetches from the
// wrapped repository and stores the result. DB errors are never cached.
func (r *CachedRepository) List(ctx context.Context, limit, offset int, userID string) ([]*residentlocationdomain.ResidentLocation, error) {
	if r.listCache == nil {
		return r.repo.List(ctx, limit, offset, userID)
	}

	key := listCacheKey(r.generation(userID), userID, limit, offset)
	if cached, ok := r.listCache.Get(key); ok {
		return cloneLocations(cached), nil
	}

	items, err := r.repo.List(ctx, limit, offset, userID)
	if err != nil {
		return nil, err
	}

	r.listCache.Add(key, cloneLocations(items))
	return items, nil
}

// Count returns a cached total when available; on a miss it fetches from the
// wrapped repository and stores the result. DB errors are never cached.
func (r *CachedRepository) Count(ctx context.Context, userID string) (int, error) {
	if r.countCache == nil {
		return r.repo.Count(ctx, userID)
	}

	key := countCacheKey(r.generation(userID), userID)
	if cached, ok := r.countCache.Get(key); ok {
		return cached, nil
	}

	count, err := r.repo.Count(ctx, userID)
	if err != nil {
		return 0, err
	}

	r.countCache.Add(key, count)
	return count, nil
}

// Create writes through to the wrapped repository, then bumps the user's
// generation counter on success (invalidating only that user's cached entries).
func (r *CachedRepository) Create(ctx context.Context, input *residentlocationdomain.ResidentLocationInput, userID string) (*residentlocationdomain.ResidentLocation, error) {
	location, err := r.repo.Create(ctx, input, userID)
	if err != nil {
		return nil, err
	}
	r.bump(userID)
	return location, nil
}

// Update writes through to the wrapped repository, then bumps the user's
// generation counter on success (invalidating only that user's cached entries).
func (r *CachedRepository) Update(ctx context.Context, id int64, input *residentlocationdomain.ResidentLocationInput, userID string) (*residentlocationdomain.ResidentLocation, error) {
	location, err := r.repo.Update(ctx, id, input, userID)
	if err != nil {
		return nil, err
	}
	r.bump(userID)
	return location, nil
}

// Delete writes through to the wrapped repository, then bumps the user's
// generation counter on success (invalidating only that user's cached entries).
func (r *CachedRepository) Delete(ctx context.Context, id int64, userID string) (int64, error) {
	deleted, err := r.repo.Delete(ctx, id, userID)
	if err != nil {
		return deleted, err
	}
	r.bump(userID)
	return deleted, nil
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

// listCacheKey embeds the user's generation counter so any write by that user
// invalidates every cached page without a prefix scan.
func listCacheKey(gen int64, userID string, limit, offset int) string {
	return fmt.Sprintf("resident-location:list:%s:%d:%d:%d", userID, gen, limit, offset)
}

// countCacheKey embeds the user's generation counter; Count takes no pagination
// args.
func countCacheKey(gen int64, userID string) string {
	return fmt.Sprintf("resident-location:count:%s:%d", userID, gen)
}

// cloneLocations returns a shallow copy of the slice. *ResidentLocation values
// are immutable (private fields, getters only), so copying the slice header is
// enough to stop callers from mutating the cached backing array via
// append/slicing.
func cloneLocations(items []*residentlocationdomain.ResidentLocation) []*residentlocationdomain.ResidentLocation {
	if items == nil {
		return nil
	}
	cloned := make([]*residentlocationdomain.ResidentLocation, len(items))
	copy(cloned, items)
	return cloned
}

// Compile-time check: CachedRepository satisfies residentlocationdomain.Repository.
var _ residentlocationdomain.Repository = (*CachedRepository)(nil)
