import type { IAppConfig } from '@flatty-budget/shared';

export function reactSettings(): IAppConfig {
  return {
    bundleName: 'settings',
    cssBundleName: 'settings',
    remoteOrigin: 'http://localhost:5174',
    proxyBasePath: '/external-settings',
  };
}
