import * as microApps from "#/applications";

export type MicroAppConfig = {
  bundleName: string;
  cssBundleName?: string;
  remoteOrigin?: string;
  basePath?: string;
  proxyBasePath?: string;
};

export type MicroAppFactory = () => MicroAppConfig;

export type AppFactories = typeof microApps;
export type AppConfigUnion = ReturnType<AppFactories[keyof AppFactories]>;
export type BundleName = AppConfigUnion["bundleName"];
export type AvailableConfigs = {
  [K in BundleName]: Extract<AppConfigUnion, { bundleName: K }>;
};
