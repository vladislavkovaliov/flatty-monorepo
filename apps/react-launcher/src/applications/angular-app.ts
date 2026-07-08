import type { IAppConfig } from '@flatty-budget/shared';

export function angularApp(): IAppConfig {
  return {
    bundleName: 'angular-app',
    cssBundleName: 'angular-app',
    remoteOrigin: 'http://localhost:4200',
    proxyBasePath: '/external-app',
  };
}
