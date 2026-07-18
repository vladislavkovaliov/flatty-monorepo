import { describe, it, expect } from 'vitest';
import { USERS_QUERIES } from './users.queries';

describe('USERS_QUERIES', () => {
  it('has a valid all() query key', () => {
    const key = USERS_QUERIES.all();
    expect(key).toEqual(['users']);
  });

  it('list() has correct query key shape and queryFn', () => {
    const opts = USERS_QUERIES.list(15, 0);
    expect(opts.queryKey).toEqual(['users', 'list', { limit: 15, offset: 0 }]);
    expect(typeof opts.queryFn).toBe('function');
  });

  it('byId() has correct query key shape and queryFn', () => {
    const opts = USERS_QUERIES.byId('abc-123');
    expect(opts.queryKey).toEqual(['users', 'byId', 'abc-123']);
    expect(typeof opts.queryFn).toBe('function');
  });
});
