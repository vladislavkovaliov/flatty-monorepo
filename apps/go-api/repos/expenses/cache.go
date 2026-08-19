package expenses

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/hashicorp/golang-lru/v2/expirable"

	expensesdomain "flatty-budget/go-api/domains/expenses"
)

// Cache defaults for the CachedRepository decorator.
const (
	defaultCacheSize = 1024
	defaultCacheTTL  = time.Hour
)

// CachedRepository decorates an expensesdomain.Repository with an in-memory TTL
// cache for Count/List/GetYearsAndMonths/GetExpensesByYearMonth results. Write
// operations (Create/Update/Delete) bump a per-user generation counter embedded
// in every cache key, which invalidates only that user's cached entries on the
// next read. Thread-safe via the underlying LRUs and a mutex-guarded generation
// map.
//
// GetByID is intentionally NOT cached: it runs immediately before writes to
// produce the Kafka pre-image event and is not user-scoped.
//
// Fail-open: if the caches are nil (or an entry is absent) the decorator simply
// delegates to the wrapped repository — a cache problem can never break the
// request path. DB errors are never cached.
type CachedRepository struct {
	repo             expensesdomain.Repository
	listCache        *expirable.LRU[string, []*expensesdomain.ExpenseWithCategory]
	countCache       *expirable.LRU[string, int]
	yearsMonthsCache *expirable.LRU[string, []*expensesdomain.YearAndMonth]
	mu               sync.Mutex
	gen              map[string]int64
}

// NewCachedRepository wraps repo with a default in-memory LRU cache (1024
// entries, 1 hour TTL).
func NewCachedRepository(repo expensesdomain.Repository) *CachedRepository {
	return NewCachedRepositoryWithTTL(repo, defaultCacheTTL)
}

// NewCachedRepositoryWithTTL wraps repo with an in-memory LRU cache using the
// given TTL. Exposed for tests (tiny TTLs) — production code uses
// NewCachedRepository.
func NewCachedRepositoryWithTTL(repo expensesdomain.Repository, ttl time.Duration) *CachedRepository {
	return &CachedRepository{
		repo:             repo,
		listCache:        expirable.NewLRU[string, []*expensesdomain.ExpenseWithCategory](defaultCacheSize, nil, ttl),
		countCache:       expirable.NewLRU[string, int](defaultCacheSize, nil, ttl),
		yearsMonthsCache: expirable.NewLRU[string, []*expensesdomain.YearAndMonth](defaultCacheSize, nil, ttl),
	}
}

// List returns a cached page when available; on a miss it fetches from the
// wrapped repository and stores the result. DB errors are never cached.
func (r *CachedRepository) List(ctx context.Context, residentLocationID int64, userID string, limit, offset int) ([]*expensesdomain.ExpenseWithCategory, error) {
	if r.listCache == nil {
		return r.repo.List(ctx, residentLocationID, userID, limit, offset)
	}

	key := listCacheKey(r.generation(userID), residentLocationID, userID, limit, offset)
	if cached, ok := r.listCache.Get(key); ok {
		return cloneExpenses(cached), nil
	}

	items, err := r.repo.List(ctx, residentLocationID, userID, limit, offset)
	if err != nil {
		return nil, err
	}

	r.listCache.Add(key, cloneExpenses(items))
	return items, nil
}

// Count returns a cached total when available; on a miss it fetches from the
// wrapped repository and stores the result. DB errors are never cached.
func (r *CachedRepository) Count(ctx context.Context, residentLocationID int64, userID string) (int, error) {
	if r.countCache == nil {
		return r.repo.Count(ctx, residentLocationID, userID)
	}

	key := countCacheKey(r.generation(userID), residentLocationID, userID)
	if cached, ok := r.countCache.Get(key); ok {
		return cached, nil
	}

	count, err := r.repo.Count(ctx, residentLocationID, userID)
	if err != nil {
		return 0, err
	}

	r.countCache.Add(key, count)
	return count, nil
}

// GetYearsAndMonths returns a cached list when available; on a miss it fetches
// from the wrapped repository and stores the result. DB errors are never
// cached.
func (r *CachedRepository) GetYearsAndMonths(ctx context.Context, residentLocationID int64, userID string) ([]*expensesdomain.YearAndMonth, error) {
	if r.yearsMonthsCache == nil {
		return r.repo.GetYearsAndMonths(ctx, residentLocationID, userID)
	}

	key := yearsMonthsCacheKey(r.generation(userID), residentLocationID, userID)
	if cached, ok := r.yearsMonthsCache.Get(key); ok {
		return cloneYearsAndMonths(cached), nil
	}

	items, err := r.repo.GetYearsAndMonths(ctx, residentLocationID, userID)
	if err != nil {
		return nil, err
	}

	r.yearsMonthsCache.Add(key, cloneYearsAndMonths(items))
	return items, nil
}

// GetExpensesByYearMonth returns a cached list when available; on a miss it
// fetches from the wrapped repository and stores the result. Shares the list
// cache with List (different key prefixes keep entries separate). DB errors are
// never cached.
func (r *CachedRepository) GetExpensesByYearMonth(ctx context.Context, residentLocationID, year, month int64, userID string) ([]*expensesdomain.ExpenseWithCategory, error) {
	if r.listCache == nil {
		return r.repo.GetExpensesByYearMonth(ctx, residentLocationID, year, month, userID)
	}

	key := byYearMonthCacheKey(r.generation(userID), residentLocationID, userID, year, month)
	if cached, ok := r.listCache.Get(key); ok {
		return cloneExpenses(cached), nil
	}

	items, err := r.repo.GetExpensesByYearMonth(ctx, residentLocationID, year, month, userID)
	if err != nil {
		return nil, err
	}

	r.listCache.Add(key, cloneExpenses(items))
	return items, nil
}

// Create writes through to the wrapped repository, then bumps the user's
// generation counter on success (invalidating only that user's cached entries).
func (r *CachedRepository) Create(ctx context.Context, input *expensesdomain.ExpenseInput, userID string) (*expensesdomain.Expense, error) {
	expense, err := r.repo.Create(ctx, input, userID)
	if err != nil {
		return nil, err
	}
	r.bump(userID)
	return expense, nil
}

// Update writes through to the wrapped repository, then bumps the user's
// generation counter on success (invalidating only that user's cached entries).
func (r *CachedRepository) Update(ctx context.Context, id int64, input *expensesdomain.ExpenseInput, userID string) (*expensesdomain.Expense, error) {
	expense, err := r.repo.Update(ctx, id, input, userID)
	if err != nil {
		return nil, err
	}
	r.bump(userID)
	return expense, nil
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

// GetByID is intentionally NOT cached: it runs immediately before writes to
// produce the Kafka pre-image event and is not user-scoped. It always delegates
// to the wrapped repository.
func (r *CachedRepository) GetByID(ctx context.Context, id int64) (*expensesdomain.Expense, error) {
	return r.repo.GetByID(ctx, id)
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
func listCacheKey(gen int64, residentLocationID int64, userID string, limit, offset int) string {
	return fmt.Sprintf("expenses:list:%s:%d:%d:%d:%d", userID, residentLocationID, gen, limit, offset)
}

// countCacheKey embeds the user's generation counter; Count takes no pagination
// args.
func countCacheKey(gen int64, residentLocationID int64, userID string) string {
	return fmt.Sprintf("expenses:count:%s:%d:%d", userID, residentLocationID, gen)
}

// yearsMonthsCacheKey embeds the user's generation counter.
func yearsMonthsCacheKey(gen int64, residentLocationID int64, userID string) string {
	return fmt.Sprintf("expenses:years-months:%s:%d:%d", userID, residentLocationID, gen)
}

// byYearMonthCacheKey embeds the user's generation counter plus the year/month
// filter.
func byYearMonthCacheKey(gen int64, residentLocationID int64, userID string, year, month int64) string {
	return fmt.Sprintf("expenses:by-year-month:%s:%d:%d:%d:%d", userID, residentLocationID, gen, year, month)
}

// cloneExpenses returns a shallow copy of the slice. *ExpenseWithCategory
// values are immutable (private fields, getters only), so copying the slice
// header is enough to stop callers from mutating the cached backing array via
// append/slicing.
func cloneExpenses(items []*expensesdomain.ExpenseWithCategory) []*expensesdomain.ExpenseWithCategory {
	if items == nil {
		return nil
	}
	cloned := make([]*expensesdomain.ExpenseWithCategory, len(items))
	copy(cloned, items)
	return cloned
}

// cloneYearsAndMonths returns a shallow copy of the slice. *YearAndMonth values
// are immutable (private fields, getters only).
func cloneYearsAndMonths(items []*expensesdomain.YearAndMonth) []*expensesdomain.YearAndMonth {
	if items == nil {
		return nil
	}
	cloned := make([]*expensesdomain.YearAndMonth, len(items))
	copy(cloned, items)
	return cloned
}

// Compile-time check: CachedRepository satisfies expensesdomain.Repository.
var _ expensesdomain.Repository = (*CachedRepository)(nil)
