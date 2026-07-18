import { describe, it, expect } from 'vitest';
import { CATEGORIES_QUERIES } from './categories.queries';

describe('CATEGORIES_QUERIES', () => {
  it('has a valid all() query key', () => {
    const key = CATEGORIES_QUERIES.all();
    expect(key).toEqual(['categories']);
  });

  it('list() has correct query key shape and queryFn', () => {
    const opts = CATEGORIES_QUERIES.list(10, 0);
    expect(opts.queryKey).toEqual(['categories', 'list', { limit: 10, offset: 0 }]);
    expect(typeof opts.queryFn).toBe('function');
  });
});
