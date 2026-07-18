// Contracts — public API
export type { IAppConfig, IMicroAppConfig, IAppComponent } from './contracts/app-config.type';
export { APP_NAMESPACES, APPS_VENDORS_CHUNK_NAME } from './contracts/app-namespaces.const';

// Remote — public exports
export { AppComponent } from './remote/app-component';
export { sanitizeConfig } from './remote/sanitize-config';

// Package-internal (NOT re-exported):
// isRecord, isAllowedEnv, isAllowedHostType, isNavigate,
// ALLOWED_ENVS, ALLOWED_HOST_TYPES, AllowedEnvType, AllowedHostType
