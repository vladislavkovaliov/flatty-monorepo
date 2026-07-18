import type { IAppConfig as BaseIAppConfig, IAppComponent, IMicroAppConfig } from '@flatty-budget/mfe-host';
import { APP_NAMESPACES, APPS_VENDORS_CHUNK_NAME } from '@flatty-budget/mfe-host';

export interface IAppConfig extends BaseIAppConfig {
    basePath?: string;
    proxyBasePath?: string;
}

export type { IAppComponent, IMicroAppConfig };
export { APP_NAMESPACES, APPS_VENDORS_CHUNK_NAME };
