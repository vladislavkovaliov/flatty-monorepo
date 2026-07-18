import { describe, it, expect } from 'vitest';

describe('@flatty-budget/mfe-host public API', () => {
  it('exports AppComponent', async () => {
    const mod = await import('./index');
    expect(mod.AppComponent).toBeDefined();
    expect(typeof mod.AppComponent).toBe('function');
  });

  it('exports sanitizeConfig', async () => {
    const mod = await import('./index');
    expect(mod.sanitizeConfig).toBeDefined();
    expect(typeof mod.sanitizeConfig).toBe('function');
  });

  it('exports contract types', async () => {
    const mod = await import('./index');
    expect(mod.APP_NAMESPACES).toBe('ext-apps');
    expect(mod.APPS_VENDORS_CHUNK_NAME).toBe('apps');
  });

  it('does NOT export internal helpers', async () => {
    const mod = await import('./index');
    expect(mod).not.toHaveProperty('isRecord');
    expect(mod).not.toHaveProperty('isAllowedEnv');
    expect(mod).not.toHaveProperty('isAllowedHostType');
    expect(mod).not.toHaveProperty('isNavigate');
    expect(mod).not.toHaveProperty('ALLOWED_ENVS');
    expect(mod).not.toHaveProperty('ALLOWED_HOST_TYPES');
  });
});
