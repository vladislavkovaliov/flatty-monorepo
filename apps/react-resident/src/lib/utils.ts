import type {IAppConfig} from '../types/external-app-config.type';

export const ALLOWED_ENVS = ['development', 'production', 'qa'] as const;
export const ALLOWED_HOST_TYPES = ['angular', 'react', 'other'] as const;
export type AllowedEnvType = typeof ALLOWED_ENVS[number];
export type AllowedHostType = typeof ALLOWED_HOST_TYPES[number];

export function isRecord(value: unknown): value is Record<string, unknown> {
    return typeof value === 'object' && value !== null && !Array.isArray(value);
}

export function isAllowedEnv(value: string): value is AllowedEnvType {
    return ALLOWED_ENVS.includes(value as AllowedEnvType);
}

export function isAllowedHostType(value: string): value is AllowedHostType {
    return ALLOWED_HOST_TYPES.includes(value as AllowedHostType);
}

export function isNavigate(value: unknown): value is () => Promise<void> {
    return typeof value === 'function';
}

export function sanitizeConfig(config: unknown): IAppConfig {
    const safeConfig: IAppConfig = {
        env: 'production',
        featureFlags: {},
        hostType: 'other',
    };

    if (!config || typeof config !== 'object') {
        return safeConfig;
    }

    const cfg = config as Record<string, unknown>;

    if (typeof cfg.env === 'string' && isAllowedEnv(cfg.env)) {
        safeConfig.env = cfg.env;
    }

    if (typeof cfg.hostType === 'string' && isAllowedHostType(cfg.hostType)) {
        safeConfig.hostType = cfg.hostType;
    }

    if (isRecord(cfg.featureFlags)) {
        safeConfig.featureFlags = cfg.featureFlags;
    }

    if (isNavigate(cfg.navigate)) {
        safeConfig.navigate = cfg.navigate;
    }

    if (typeof cfg.iconsSpriteUrl === 'string' && cfg.iconsSpriteUrl.startsWith('http')) {
        safeConfig.iconsSpriteUrl = cfg.iconsSpriteUrl;
    }

    return safeConfig;
}

async function requestJson<T>(url: string, options: RequestInit): Promise<T> {
  const response = await fetch(url, options);

  if (!response.ok) {
    const body = await response.json().catch(() => null);
    throw new Error(body?.error ?? `Request failed (${response.status})`);
  }

  return response.json() as Promise<T>;
}

export async function getJson<T>(url: string): Promise<T> {
  return requestJson<T>(url, { method: 'GET' });
}

export async function postJson<T>(url: string, body: unknown): Promise<T> {
  return requestJson<T>(url, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(body),
  });
}

export async function putJson<T>(url: string, body: unknown): Promise<T> {
  return requestJson<T>(url, {
    method: 'PUT',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(body),
  });
}

export async function deleteJson<T>(url: string): Promise<T> {
  return requestJson<T>(url, { method: 'DELETE' });
}