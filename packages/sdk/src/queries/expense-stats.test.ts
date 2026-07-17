import { describe, it, expect } from 'vitest';
import { EXPENSE_STATS_GRAPHQL_QUERIES } from './expense-stats.graphql';

describe('EXPENSE_STATS_GRAPHQL_QUERIES', () => {
  it('has a valid all() query key', () => {
    const key = EXPENSE_STATS_GRAPHQL_QUERIES.all();
    expect(key).toEqual(['expense-stats', 'graphql']);
  });

  it('totals() has correct query key shape', () => {
    const opts = EXPENSE_STATS_GRAPHQL_QUERIES.totals();
    expect(opts.queryKey).toEqual(['expense-stats', 'graphql', 'totals']);
    expect(typeof opts.queryFn).toBe('function');
  });

  it('averages() has correct query key shape', () => {
    const opts = EXPENSE_STATS_GRAPHQL_QUERIES.averages();
    expect(opts.queryKey).toEqual(['expense-stats', 'graphql', 'averages']);
    expect(typeof opts.queryFn).toBe('function');
  });
});
