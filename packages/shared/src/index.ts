export interface IAppConfig {
  bundleName: string;
  cssBundleName?: string;
  remoteOrigin: string;
  proxyBasePath: string;
}

export interface IAppComponent {
  initialize(element: HTMLElement, config: Record<string, unknown>): void;
  destroy(element?: HTMLElement): void;
}
