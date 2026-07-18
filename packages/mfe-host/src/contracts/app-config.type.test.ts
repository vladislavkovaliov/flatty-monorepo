import { describe, it, expect, expectTypeOf } from 'vitest';
import type { IAppConfig, IMicroAppConfig, IAppComponent } from './app-config.type';

describe('IAppConfig', () => {
  it('has base fields', () => {
    const config: IAppConfig = {
      env: 'development',
      featureFlags: {},
      hostType: 'react',
    };
    expectTypeOf(config).toMatchTypeOf<{
      bundleName?: string;
      cssBundleName?: string;
      remoteOrigin?: string;
      env: string;
      featureFlags: Record<string, unknown>;
      navigate?: () => Promise<void>;
      iconsSpriteUrl?: string;
      hostType: 'angular' | 'react' | 'other';
    }>();
  });

  it('allows optional fields to be omitted', () => {
    const minimal: IAppConfig = {
      env: 'production',
      featureFlags: {},
      hostType: 'other',
    };
    expect(minimal.env).toBe('production');
  });
});

describe('IMicroAppConfig', () => {
  it('has required bundle fields', () => {
    const config: IMicroAppConfig = {
      bundleName: 'app',
      cssBundleName: 'styles',
      remoteOrigin: 'http://localhost:8081',
    };
    expectTypeOf(config).toMatchTypeOf<{
      bundleName: string;
      cssBundleName: string;
      remoteOrigin: string;
    }>();
  });
});

describe('IAppComponent', () => {
  it('has initialize and destroy methods', () => {
    const component: IAppComponent = {
      initialize(_element: Element, _config: IAppConfig) {},
      destroy() {},
    };
    expectTypeOf(component.initialize).toBeFunction();
    expectTypeOf(component.destroy).toBeFunction();
  });
});
