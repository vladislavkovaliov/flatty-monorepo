package category

import (
	"context"
	"fmt"
	"sync/atomic"
	"time"

	"github.com/hashicorp/golang-lru/v2/expirable"

	categorydomain "flatty-budget/go-api/domains/category"
)

// Cache defaults for the CachedRepository decorator.
const (
	defaultCacheSize = 1024
	defaultCacheTTL  = time.Hour
)

// CachedRepository decorates a categorydomain.Repository with an in-memory TTL
// cache for List/Count results. Write operations (Create/Update/Delete) bump a
// generation counter embedded in every cache key, which invalidates all cached
// list/count entries on the next read. Thread-safe via the underlying LRU and
// atomic counter.
//
// Fail-open: if the caches are nil (or an entry is absent) the decorator simply
// delegates to the wrapped repository — a cache problem can never break the
// request path. DB errors are never cached.
type CachedRepository struct {
	repo       categorydomain.Repository
	listCache  *expirable.LRU[string, []*categorydomain.Category]
	countCache *expirable.LRU[string, int]
	gen        atomic.Int64
}

// NewCachedRepository wraps repo with a default in-memory LRU cache (1024
// entries, 1 hour TTL).
func NewCachedRepository(repo categorydomain.Repository) *CachedRepository {
	return NewCachedRepositoryWithTTL(repo, defaultCacheTTL)
}

// NewCachedRepositoryWithTTL wraps repo with an in-memory LRU cache using the
// given TTL. Exposed for tests (tiny TTLs) — production code uses
// NewCachedRepository.
func NewCachedRepositoryWithTTL(repo categorydomain.Repository, ttl time.Duration) *CachedRepository {
	return &CachedRepository{
		repo:       repo,
		listCache:  expirable.NewLRU[string, []*categorydomain.Category](defaultCacheSize, nil, ttl),
		countCache: expirable.NewLRU[string, int](defaultCacheSize, nil, ttl),
	}
}

// List returns a cached page when available; on a miss it fetches from the
// wrapped repository and stores the result. DB errors are never cached.
func (r *CachedRepository) List(ctx context.Context, limit, offset int) ([]*categorydomain.Category, error) {
	if r.listCache == nil {
		return r.repo.List(ctx, limit, offset)
	}

	key := listCacheKey(r.gen.Load(), limit, offset)
	if cached, ok := r.listCache.Get(key); ok {
		return cloneCategories(cached), nil
	}

	items, err := r.repo.List(ctx, limit, offset)
	if err != nil {
		return nil, err
	}

	r.listCache.Add(key, cloneCategories(items))
	return items, nil
}

// Count returns a cached total when available; on a miss it fetches from the
// wrapped repository and stores the result. DB errors are never cached.
func (r *CachedRepository) Count(ctx context.Context) (int, error) {
	if r.countCache == nil {
		return r.repo.Count(ctx)
	}

	key := countCacheKey(r.gen.Load())
	if cached, ok := r.countCache.Get(key); ok {
		return cached, nil
	}

	count, err := r.repo.Count(ctx)
	if err != nil {
		return 0, err
	}

	r.countCache.Add(key, count)
	return count, nil
}

// Create writes through to the wrapped repository, then bumps the generation
// counter on success (invalidating all cached list/count entries).
func (r *CachedRepository) Create(ctx context.Context, input *categorydomain.CategoryInput) (*categorydomain.Category, error) {
	category, err := r.repo.Create(ctx, input)
	if err != nil {
		return nil, err
	}
	r.gen.Add(1)
	return category, nil
}

// Update writes through to the wrapped repository, then bumps the generation
// counter on success (invalidating all cached list/count entries).
func (r *CachedRepository) Update(ctx context.Context, id int64, input *categorydomain.CategoryInput) (*categorydomain.Category, error) {
	category, err := r.repo.Update(ctx, id, input)
	if err != nil {
		return nil, err
	}
	r.gen.Add(1)
	return category, nil
}

// Delete writes through to the wrapped repository, then bumps the generation
// counter on success (invalidating all cached list/count entries).
func (r *CachedRepository) Delete(ctx context.Context, id int64) (int64, error) {
	deleted, err := r.repo.Delete(ctx, id)
	if err != nil {
		return deleted, err
	}
	r.gen.Add(1)
	return deleted, nil
}

// listCacheKey embeds the generation counter so any write invalidates every
// cached page key without a prefix scan (unsupported by TTL caches).
func listCacheKey(gen int64, limit, offset int) string {
	return fmt.Sprintf("categories:list:%d:%d:%d", gen, limit, offset)
}

// countCacheKey embeds the generation counter; Count takes no pagination args.
func countCacheKey(gen int64) string {
	return fmt.Sprintf("categories:count:%d", gen)
}

// cloneCategories returns a shallow copy of the slice. *Category values are
// immutable (private fields, getters only — domains/category/category.go), so
// copying the slice header is enough to stop callers from mutating the cached
// backing array via append/slicing.
func cloneCategories(categories []*categorydomain.Category) []*categorydomain.Category {
	if categories == nil {
		return nil
	}
	cloned := make([]*categorydomain.Category, len(categories))
	copy(cloned, categories)
	return cloned
}

// Compile-time check: CachedRepository satisfies categorydomain.Repository.
var _ categorydomain.Repository = (*CachedRepository)(nil)
