import { describe, it, expect } from 'vitest';
import { RESIDENT_LOCATION_QUERIES } from './resident-location.queries';

describe('RESIDENT_LOCATION_QUERIES', () => {
  it('has a valid all() query key', () => {
    const key = RESIDENT_LOCATION_QUERIES.all();
    expect(key).toEqual(['resident-location']);
  });

  it('current() has correct query key shape and queryFn', () => {
    const opts = RESIDENT_LOCATION_QUERIES.current();
    expect(opts.queryKey).toEqual(['resident-location', 'current']);
    expect(typeof opts.queryFn).toBe('function');
  });
});
