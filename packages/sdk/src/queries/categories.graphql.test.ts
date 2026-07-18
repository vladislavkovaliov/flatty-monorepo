import { describe, it, expect } from 'vitest';
import { CATEGORIES_GRAPHQL_QUERIES } from './categories.graphql';

describe('CATEGORIES_GRAPHQL_QUERIES', () => {
  it('has a valid all() query key', () => {
    const key = CATEGORIES_GRAPHQL_QUERIES.all();
    expect(key).toEqual(['categories', 'graphql']);
  });

  it('list() has correct query key shape and queryFn', () => {
    const opts = CATEGORIES_GRAPHQL_QUERIES.list(5, 10);
    expect(opts.queryKey).toEqual(['categories', 'graphql', 'list', { limit: 5, offset: 10 }]);
    expect(typeof opts.queryFn).toBe('function');
  });
});
