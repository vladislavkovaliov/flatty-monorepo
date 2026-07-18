import { describe, it, expect } from 'vitest';
import { USERS_GRAPHQL_QUERIES } from './users.graphql';

describe('USERS_GRAPHQL_QUERIES', () => {
  it('has a valid all() query key', () => {
    const key = USERS_GRAPHQL_QUERIES.all();
    expect(key).toEqual(['users', 'graphql']);
  });

  it('list() has correct query key shape and queryFn', () => {
    const opts = USERS_GRAPHQL_QUERIES.list(10, 0);
    expect(opts.queryKey).toEqual(['users', 'graphql', 'list', { limit: 10, offset: 0 }]);
    expect(typeof opts.queryFn).toBe('function');
  });

  it('byId() has correct query key shape and queryFn', () => {
    const opts = USERS_GRAPHQL_QUERIES.byId('user-1');
    expect(opts.queryKey).toEqual(['users', 'graphql', 'byId', 'user-1']);
    expect(typeof opts.queryFn).toBe('function');
  });
});
