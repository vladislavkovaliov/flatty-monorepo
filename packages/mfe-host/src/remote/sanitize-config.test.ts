import { describe, it, expect } from 'vitest';
import { sanitizeConfig } from './sanitize-config';

describe('sanitizeConfig', () => {
  it('returns safe defaults for null', () => {
    const result = sanitizeConfig(null);
    expect(result).toEqual({
      env: 'production',
      featureFlags: {},
      hostType: 'other',
    });
  });

  it('returns safe defaults for undefined', () => {
    const result = sanitizeConfig(undefined);
    expect(result).toEqual({
      env: 'production',
      featureFlags: {},
      hostType: 'other',
    });
  });

  it('returns safe defaults for non-object', () => {
    const result = sanitizeConfig('invalid');
    expect(result).toEqual({
      env: 'production',
      featureFlags: {},
      hostType: 'other',
    });
  });

  it('preserves valid env', () => {
    const result = sanitizeConfig({ env: 'development' });
    expect(result.env).toBe('development');
  });

  it('preserves qa env', () => {
    const result = sanitizeConfig({ env: 'qa' });
    expect(result.env).toBe('qa');
  });

  it('preserves production env', () => {
    const result = sanitizeConfig({ env: 'production' });
    expect(result.env).toBe('production');
  });

  it('rejects invalid env and falls back to production', () => {
    const result = sanitizeConfig({ env: 'staging' });
    expect(result.env).toBe('production');
  });

  it('preserves valid hostType', () => {
    const result = sanitizeConfig({ hostType: 'angular' });
    expect(result.hostType).toBe('angular');
  });

  it('rejects invalid hostType', () => {
    const result = sanitizeConfig({ hostType: 'vue' });
    expect(result.hostType).toBe('other');
  });

  it('preserves valid featureFlags as record', () => {
    const flags = { featureX: true, count: 42 };
    const result = sanitizeConfig({ featureFlags: flags });
    expect(result.featureFlags).toEqual(flags);
  });

  it('rejects non-record featureFlags (array)', () => {
    const result = sanitizeConfig({ featureFlags: ['a', 'b'] });
    expect(result.featureFlags).toEqual({});
  });

  it('rejects non-record featureFlags (string)', () => {
    const result = sanitizeConfig({ featureFlags: 'invalid' });
    expect(result.featureFlags).toEqual({});
  });

  it('preserves navigate function', () => {
    const navigate = async () => {};
    const result = sanitizeConfig({ navigate });
    expect(result.navigate).toBe(navigate);
  });

  it('strips non-function navigate', () => {
    const result = sanitizeConfig({ navigate: 'not-a-function' });
    expect(result.navigate).toBeUndefined();
  });

  it('preserves valid http iconsSpriteUrl', () => {
    const result = sanitizeConfig({ iconsSpriteUrl: 'https://example.com/sprite.svg' });
    expect(result.iconsSpriteUrl).toBe('https://example.com/sprite.svg');
  });

  it('preserves valid http (non-https) iconsSpriteUrl', () => {
    const result = sanitizeConfig({ iconsSpriteUrl: 'http://example.com/sprite.svg' });
    expect(result.iconsSpriteUrl).toBe('http://example.com/sprite.svg');
  });

  it('rejects non-http iconsSpriteUrl', () => {
    const result = sanitizeConfig({ iconsSpriteUrl: 'ftp://example.com/sprite.svg' });
    expect(result.iconsSpriteUrl).toBeUndefined();
  });

  it('rejects iconsSpriteUrl that is not a string', () => {
    const result = sanitizeConfig({ iconsSpriteUrl: 123 });
    expect(result.iconsSpriteUrl).toBeUndefined();
  });

  it('handles empty object', () => {
    const result = sanitizeConfig({});
    expect(result).toEqual({
      env: 'production',
      featureFlags: {},
      hostType: 'other',
    });
  });
});
