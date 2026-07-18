export interface IAppConfig {
    bundleName?: string;
    cssBundleName?: string;
    remoteOrigin?: string;
    env: string;
    featureFlags: Record<string, unknown>;
    navigate?: () => Promise<void>;
    iconsSpriteUrl?: string;
    hostType: 'angular' | 'react' | 'other';
}

export interface IMicroAppConfig {
    bundleName: string;
    cssBundleName: string;
    remoteOrigin: string;
}

export interface IAppComponent {
    initialize(element: Element, _config: IAppConfig): void;

    destroy: () => void;
}

export const APP_NAMESPACES = 'ext-apps';

export const APPS_VENDORS_CHUNK_NAME = 'apps';