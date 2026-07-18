import { describe, it, expect } from 'vitest';
import { EXPENSES_QUERIES } from './expenses.queries';

describe('EXPENSES_QUERIES', () => {
  it('has a valid all() query key', () => {
    const key = EXPENSES_QUERIES.all();
    expect(key).toEqual(['expenses']);
  });

  it('list() has correct query key shape and queryFn', () => {
    const opts = EXPENSES_QUERIES.list(20, 0);
    expect(opts.queryKey).toEqual(['expenses', 'list', { limit: 20, offset: 0 }]);
    expect(typeof opts.queryFn).toBe('function');
  });
});
