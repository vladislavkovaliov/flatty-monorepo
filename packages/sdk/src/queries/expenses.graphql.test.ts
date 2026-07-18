import { describe, it, expect } from 'vitest';
import { EXPENSES_GRAPHQL_QUERIES } from './expenses.graphql';

describe('EXPENSES_GRAPHQL_QUERIES', () => {
  it('has a valid all() query key', () => {
    const key = EXPENSES_GRAPHQL_QUERIES.all();
    expect(key).toEqual(['expenses', 'graphql']);
  });

  it('list() has correct query key shape and queryFn', () => {
    const opts = EXPENSES_GRAPHQL_QUERIES.list(5, 0);
    expect(opts.queryKey).toEqual(['expenses', 'graphql', 'list', { limit: 5, offset: 0 }]);
    expect(typeof opts.queryFn).toBe('function');
  });
});
