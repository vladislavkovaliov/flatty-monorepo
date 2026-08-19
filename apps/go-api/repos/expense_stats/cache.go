package expense_stats

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/hashicorp/golang-lru/v2/expirable"

	expensestatsdomain "flatty-budget/go-api/domains/expense_stats"
)

// Cache defaults for the cached repository decorators.
const (
	defaultCacheSize = 1024
	// defaultStatsTTL is short because production writes flow through the Kafka
	// consumer, which the repository cannot observe — the TTL bounds staleness.
	defaultStatsTTL = time.Minute
)

// CachedMonthlyTotalRepository decorates a MonthlyTotalRepository with an
// in-memory TTL cache for List results. UpsertTotal is a pass-through: writes
// arrive via the Kafka consumer (not through this repo), so there is no
// generation counter to invalidate — the short TTL bounds staleness instead.
//
// Fail-open: if the cache is nil (or an entry is absent) the decorator simply
// delegates to the wrapped repository. DB errors are never cached.
type CachedMonthlyTotalRepository struct {
	repo      expensestatsdomain.MonthlyTotalRepository
	listCache *expirable.LRU[string, []*expensestatsdomain.ExpenseMonthlyTotal]
}

// NewCachedMonthlyTotalRepository wraps repo with a default in-memory LRU cache
// (1024 entries, 1 minute TTL).
func NewCachedMonthlyTotalRepository(repo expensestatsdomain.MonthlyTotalRepository) *CachedMonthlyTotalRepository {
	return NewCachedMonthlyTotalRepositoryWithTTL(repo, defaultStatsTTL)
}

// NewCachedMonthlyTotalRepositoryWithTTL wraps repo with an in-memory LRU cache
// using the given TTL. Exposed for tests (tiny TTLs) — production code uses
// NewCachedMonthlyTotalRepository.
func NewCachedMonthlyTotalRepositoryWithTTL(repo expensestatsdomain.MonthlyTotalRepository, ttl time.Duration) *CachedMonthlyTotalRepository {
	return &CachedMonthlyTotalRepository{
		repo:      repo,
		listCache: expirable.NewLRU[string, []*expensestatsdomain.ExpenseMonthlyTotal](defaultCacheSize, nil, ttl),
	}
}

// List returns a cached list when available; on a miss it fetches from the
// wrapped repository and stores the result. DB errors are never cached.
func (r *CachedMonthlyTotalRepository) List(ctx context.Context, residentLocationID int64, userID string, month, year *int) ([]*expensestatsdomain.ExpenseMonthlyTotal, error) {
	if r.listCache == nil {
		return r.repo.List(ctx, residentLocationID, userID, month, year)
	}

	key := totalsListCacheKey(residentLocationID, userID, month, year)
	if cached, ok := r.listCache.Get(key); ok {
		return cloneMonthlyTotals(cached), nil
	}

	items, err := r.repo.List(ctx, residentLocationID, userID, month, year)
	if err != nil {
		return nil, err
	}

	r.listCache.Add(key, cloneMonthlyTotals(items))
	return items, nil
}

// UpsertTotal is a pass-through: production writes arrive via the Kafka
// consumer, which this repository cannot observe, so there is no cache to
// invalidate — the short TTL bounds staleness.
func (r *CachedMonthlyTotalRepository) UpsertTotal(ctx context.Context, residentLocationID int64, month, year int, totalSpent float64) error {
	return r.repo.UpsertTotal(ctx, residentLocationID, month, year, totalSpent)
}

// CachedMonthlyAverageRepository decorates a MonthlyAverageRepository with an
// in-memory TTL cache for List results. UpsertAverage is a pass-through for the
// same reason as UpsertTotal (writes arrive via the Kafka consumer).
//
// Fail-open: if the cache is nil (or an entry is absent) the decorator simply
// delegates to the wrapped repository. DB errors are never cached.
type CachedMonthlyAverageRepository struct {
	repo      expensestatsdomain.MonthlyAverageRepository
	listCache *expirable.LRU[string, []*expensestatsdomain.ExpenseMonthlyAverage]
}

// NewCachedMonthlyAverageRepository wraps repo with a default in-memory LRU
// cache (1024 entries, 1 minute TTL).
func NewCachedMonthlyAverageRepository(repo expensestatsdomain.MonthlyAverageRepository) *CachedMonthlyAverageRepository {
	return NewCachedMonthlyAverageRepositoryWithTTL(repo, defaultStatsTTL)
}

// NewCachedMonthlyAverageRepositoryWithTTL wraps repo with an in-memory LRU
// cache using the given TTL. Exposed for tests (tiny TTLs) — production code
// uses NewCachedMonthlyAverageRepository.
func NewCachedMonthlyAverageRepositoryWithTTL(repo expensestatsdomain.MonthlyAverageRepository, ttl time.Duration) *CachedMonthlyAverageRepository {
	return &CachedMonthlyAverageRepository{
		repo:      repo,
		listCache: expirable.NewLRU[string, []*expensestatsdomain.ExpenseMonthlyAverage](defaultCacheSize, nil, ttl),
	}
}

// List returns a cached list when available; on a miss it fetches from the
// wrapped repository and stores the result. DB errors are never cached.
func (r *CachedMonthlyAverageRepository) List(ctx context.Context, residentLocationID int64, userID string, month, year *int) ([]*expensestatsdomain.ExpenseMonthlyAverage, error) {
	if r.listCache == nil {
		return r.repo.List(ctx, residentLocationID, userID, month, year)
	}

	key := averagesListCacheKey(residentLocationID, userID, month, year)
	if cached, ok := r.listCache.Get(key); ok {
		return cloneMonthlyAverages(cached), nil
	}

	items, err := r.repo.List(ctx, residentLocationID, userID, month, year)
	if err != nil {
		return nil, err
	}

	r.listCache.Add(key, cloneMonthlyAverages(items))
	return items, nil
}

// UpsertAverage is a pass-through: production writes arrive via the Kafka
// consumer, which this repository cannot observe, so there is no cache to
// invalidate — the short TTL bounds staleness.
func (r *CachedMonthlyAverageRepository) UpsertAverage(ctx context.Context, residentLocationID int64, month, year int, averageAmount float64, expenseCount int) error {
	return r.repo.UpsertAverage(ctx, residentLocationID, month, year, averageAmount, expenseCount)
}

// itoaOrNone renders a *int as its decimal string, or "none" when nil, so nil
// and non-nil filters get distinct cache keys.
func itoaOrNone(v *int) string {
	if v == nil {
		return "none"
	}
	return strconv.Itoa(*v)
}

// totalsListCacheKey embeds every List filter; there is no generation counter
// (writes are not observable through this repo), so staleness is bounded by the
// TTL alone.
func totalsListCacheKey(residentLocationID int64, userID string, month, year *int) string {
	return fmt.Sprintf("stats:totals:list:%d:%s:%s:%s", residentLocationID, userID, itoaOrNone(month), itoaOrNone(year))
}

// averagesListCacheKey embeds every List filter; see totalsListCacheKey.
func averagesListCacheKey(residentLocationID int64, userID string, month, year *int) string {
	return fmt.Sprintf("stats:averages:list:%d:%s:%s:%s", residentLocationID, userID, itoaOrNone(month), itoaOrNone(year))
}

// cloneMonthlyTotals returns a shallow copy of the slice. *ExpenseMonthlyTotal
// values are immutable (private fields, getters only), so copying the slice
// header is enough to stop callers from mutating the cached backing array via
// append/slicing.
func cloneMonthlyTotals(items []*expensestatsdomain.ExpenseMonthlyTotal) []*expensestatsdomain.ExpenseMonthlyTotal {
	if items == nil {
		return nil
	}
	cloned := make([]*expensestatsdomain.ExpenseMonthlyTotal, len(items))
	copy(cloned, items)
	return cloned
}

// cloneMonthlyAverages returns a shallow copy of the slice. *ExpenseMonthlyAverage
// values are immutable (private fields, getters only).
func cloneMonthlyAverages(items []*expensestatsdomain.ExpenseMonthlyAverage) []*expensestatsdomain.ExpenseMonthlyAverage {
	if items == nil {
		return nil
	}
	cloned := make([]*expensestatsdomain.ExpenseMonthlyAverage, len(items))
	copy(cloned, items)
	return cloned
}

// Compile-time checks: the decorators satisfy the domain interfaces.
var _ expensestatsdomain.MonthlyTotalRepository = (*CachedMonthlyTotalRepository)(nil)
var _ expensestatsdomain.MonthlyAverageRepository = (*CachedMonthlyAverageRepository)(nil)
