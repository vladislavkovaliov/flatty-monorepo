import { describe, it, expect } from 'vitest';
import { USER_SETTINGS_QUERIES } from './user-settings.queries';

describe('USER_SETTINGS_QUERIES', () => {
  it('has a valid all() query key', () => {
    const key = USER_SETTINGS_QUERIES.all();
    expect(key).toEqual(['user-settings']);
  });

  it('current() has correct query key shape and queryFn', () => {
    const opts = USER_SETTINGS_QUERIES.current();
    expect(opts.queryKey).toEqual(['user-settings', 'current']);
    expect(typeof opts.queryFn).toBe('function');
  });
});
