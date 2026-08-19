package applications

import (
	"context"
	"fmt"
	"sync/atomic"
	"time"

	"github.com/hashicorp/golang-lru/v2/expirable"

	applicationsdomain "flatty-budget/go-api/domains/applications"
)

// Cache defaults for the CachedRepository decorator.
const (
	defaultCacheSize = 1024
	defaultCacheTTL  = time.Hour
)

// CachedRepository decorates an applicationsdomain.Repository with an in-memory
// TTL cache for ListByEnv/ListAll results. Write operations (Create/Update/
// Delete) bump a generation counter embedded in every cache key, which
// invalidates all cached list entries on the next read. Thread-safe via the
// underlying LRU and atomic counter.
//
// Fail-open: if the caches are nil (or an entry is absent) the decorator simply
// delegates to the wrapped repository — a cache problem can never break the
// request path. DB errors are never cached.
type CachedRepository struct {
	repo           applicationsdomain.Repository
	listByEnvCache *expirable.LRU[string, []*applicationsdomain.Application]
	listAllCache   *expirable.LRU[string, []*applicationsdomain.Application]
	gen            atomic.Int64
}

// NewCachedRepository wraps repo with a default in-memory LRU cache (1024
// entries, 1 hour TTL).
func NewCachedRepository(repo applicationsdomain.Repository) *CachedRepository {
	return NewCachedRepositoryWithTTL(repo, defaultCacheTTL)
}

// NewCachedRepositoryWithTTL wraps repo with an in-memory LRU cache using the
// given TTL. Exposed for tests (tiny TTLs) — production code uses
// NewCachedRepository.
func NewCachedRepositoryWithTTL(repo applicationsdomain.Repository, ttl time.Duration) *CachedRepository {
	return &CachedRepository{
		repo:           repo,
		listByEnvCache: expirable.NewLRU[string, []*applicationsdomain.Application](defaultCacheSize, nil, ttl),
		listAllCache:   expirable.NewLRU[string, []*applicationsdomain.Application](defaultCacheSize, nil, ttl),
	}
}

// ListByEnv returns a cached list when available; on a miss it fetches from the
// wrapped repository and stores the result. DB errors are never cached.
func (r *CachedRepository) ListByEnv(ctx context.Context, env string) ([]*applicationsdomain.Application, error) {
	if r.listByEnvCache == nil {
		return r.repo.ListByEnv(ctx, env)
	}

	key := listByEnvCacheKey(r.gen.Load(), env)
	if cached, ok := r.listByEnvCache.Get(key); ok {
		return cloneApplications(cached), nil
	}

	apps, err := r.repo.ListByEnv(ctx, env)
	if err != nil {
		return nil, err
	}

	r.listByEnvCache.Add(key, cloneApplications(apps))
	return apps, nil
}

// ListAll returns a cached list when available; on a miss it fetches from the
// wrapped repository and stores the result. DB errors are never cached.
func (r *CachedRepository) ListAll(ctx context.Context) ([]*applicationsdomain.Application, error) {
	if r.listAllCache == nil {
		return r.repo.ListAll(ctx)
	}

	key := listAllCacheKey(r.gen.Load())
	if cached, ok := r.listAllCache.Get(key); ok {
		return cloneApplications(cached), nil
	}

	apps, err := r.repo.ListAll(ctx)
	if err != nil {
		return nil, err
	}

	r.listAllCache.Add(key, cloneApplications(apps))
	return apps, nil
}

// Create writes through to the wrapped repository, then bumps the generation
// counter on success (invalidating all cached list entries).
func (r *CachedRepository) Create(ctx context.Context, input *applicationsdomain.ApplicationInput) (*applicationsdomain.Application, error) {
	app, err := r.repo.Create(ctx, input)
	if err != nil {
		return nil, err
	}
	r.gen.Add(1)
	return app, nil
}

// Update writes through to the wrapped repository, then bumps the generation
// counter on success (invalidating all cached list entries).
func (r *CachedRepository) Update(ctx context.Context, id int64, input *applicationsdomain.ApplicationInput) (*applicationsdomain.Application, error) {
	app, err := r.repo.Update(ctx, id, input)
	if err != nil {
		return nil, err
	}
	r.gen.Add(1)
	return app, nil
}

// Delete writes through to the wrapped repository, then bumps the generation
// counter on success (invalidating all cached list entries).
func (r *CachedRepository) Delete(ctx context.Context, id int64) (int64, error) {
	deleted, err := r.repo.Delete(ctx, id)
	if err != nil {
		return deleted, err
	}
	r.gen.Add(1)
	return deleted, nil
}

// listByEnvCacheKey embeds the generation counter and env so any write
// invalidates every cached env list without a prefix scan.
func listByEnvCacheKey(gen int64, env string) string {
	return fmt.Sprintf("applications:list-by-env:%d:%s", gen, env)
}

// listAllCacheKey embeds the generation counter; ListAll takes no extra args.
func listAllCacheKey(gen int64) string {
	return fmt.Sprintf("applications:list-all:%d", gen)
}

// cloneApplications returns a shallow copy of the slice. *Application values
// are immutable (private fields, getters only — domains/applications/
// applications.go), so copying the slice header is enough to stop callers from
// mutating the cached backing array via append/slicing.
func cloneApplications(apps []*applicationsdomain.Application) []*applicationsdomain.Application {
	if apps == nil {
		return nil
	}
	cloned := make([]*applicationsdomain.Application, len(apps))
	copy(cloned, apps)
	return cloned
}

// Compile-time check: CachedRepository satisfies applicationsdomain.Repository.
var _ applicationsdomain.Repository = (*CachedRepository)(nil)
