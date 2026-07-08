import type { IAppComponent } from '@flatty-budget/shared';

export function loadBundle(
  namespace: string,
  bundleName: string,
  proxyBasePath: string,
): Promise<IAppComponent> {
  return new Promise((resolve, reject) => {
    const globalVar = `ext-apps.${bundleName}`;

    const existing = getGlobal<IAppComponent>(globalVar);
    if (existing) {
      resolve(existing);
      return;
    }

    const script = document.createElement('script');
    script.src = `${proxyBasePath}/${bundleName}.js`;
    script.onload = () => {
      const mod = getGlobal<IAppComponent>(globalVar);
      if (mod) {
        resolve(mod);
      } else {
        reject(new Error(`Bundle "${bundleName}" not found on window.${globalVar}`));
      }
    };
    script.onerror = () => reject(new Error(`Failed to load bundle: ${bundleName}`));
    document.body.appendChild(script);
  });
}

function getGlobal<T>(path: string): T | undefined {
  const parts = path.split('.');
  let obj: Record<string, unknown> = globalThis as Record<string, unknown>;
  for (const part of parts) {
    obj = obj[part] as Record<string, unknown>;
    if (!obj) return undefined;
  }
  return obj as T;
}
