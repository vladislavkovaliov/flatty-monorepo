export interface IAppConfig {
  env: string;
  featureFlags: Record<string, unknown>;
  hostType: 'angular' | 'react' | 'other';
  navigate?: () => Promise<void>;
  baseUrl?: string;
}

export interface IAppComponent {
  initialize(element: Element, config: IAppConfig): void;
  destroy: () => void;
}

/** Vite UMD `name` from apps/react-app/vite.config.ts */
export const APP_NAMESPACES = 'ext-apps';

export const APPS_VENDORS_CHUNK_NAME = 'apps';
