import { describe, it, expect } from 'vitest';
import { RESIDENT_LOCATION_GRAPHQL_QUERIES } from './resident-location.graphql';

describe('RESIDENT_LOCATION_GRAPHQL_QUERIES', () => {
  it('has a valid all() query key', () => {
    const key = RESIDENT_LOCATION_GRAPHQL_QUERIES.all();
    expect(key).toEqual(['resident-location', 'graphql']);
  });

  it('list() has correct query key shape and queryFn', () => {
    const opts = RESIDENT_LOCATION_GRAPHQL_QUERIES.list(10, 0);
    expect(opts.queryKey).toEqual(['resident-location', 'graphql', 'list', { limit: 10, offset: 0 }]);
    expect(typeof opts.queryFn).toBe('function');
  });
});
