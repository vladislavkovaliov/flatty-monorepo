import { createElement, useEffect, useState } from "react";
import {
  APP_NAMESPACES,
  APPS_VENDORS_CHUNK_NAME,
  type IAppComponent,
  type IAppConfig,
} from "../types/external-app-config.type";
import { LoaderUtils } from "./loader";
import { APPLICATION_BUNDLE_NAME } from "./application";

export type MicrofrontendHostProps = {
  /** Base name of the JS file (e.g. `app` → loads `app.js`). */
  bundleName: string;
  /** Base name of the CSS file (e.g. `styles` → loads `styles.css`). */
  cssBundleName: string;
  /**
   * Same-origin path proxied to the static host (see `vite.config.ts` `server.proxy`).
   * @default '/external'
   */
  proxyBasePath?: string;
  /** Remote origin when not using the primary `app` bundle path (vendor chunks). */
  remoteOrigin?: string;
  config?: Partial<IAppConfig>;
  basePath?: string;
};

const defaultConfig: Partial<IAppConfig> = {};

type ApplicationBundleName =
  (typeof APPLICATION_BUNDLE_NAME)[keyof typeof APPLICATION_BUNDLE_NAME];

function getEmbeddedApp(
  win: Window,
  applicationBundleName: ApplicationBundleName,
): IAppComponent | undefined {
  const global = win as unknown as Record<string, unknown>;

  const fromVite = global[APP_NAMESPACES];

  if (
    fromVite &&
    typeof (fromVite as IAppComponent).initialize === "function"
  ) {
    return fromVite as IAppComponent;
  }

  const legacy = global[APPS_VENDORS_CHUNK_NAME] as
    | Record<string, { default?: IAppComponent }>
    | undefined;

  return legacy?.[applicationBundleName]?.default;
}

/**
 * Loads CSS → JS → calls `window['ext-apps'].initialize(host, config)`, same flow as Angular `ExternalComponent`.
 */
export function MicrofrontendHost({
  bundleName,
  cssBundleName,
  proxyBasePath,
  basePath,
  remoteOrigin,
  config = defaultConfig,
}: MicrofrontendHostProps) {
  const [hostEl, setHostEl] = useState<HTMLElement | null>(null);

  useEffect(() => {
    if (!hostEl) {
      return;
    }
    const el = hostEl;

    let cancelled = false;

    const loadBundleScript = (): Promise<void> => {
      if (bundleName === APPLICATION_BUNDLE_NAME.APP) {
        const baseUrl = `${window.location.origin}${proxyBasePath}`;

        return LoaderUtils.loadScript({ src: `${baseUrl}/${bundleName}.js` });
      }

      if (bundleName === APPLICATION_BUNDLE_NAME.SETTINGS) {
        const baseUrl = `${window.location.origin}${proxyBasePath}`;

        return LoaderUtils.loadScript({ src: `${baseUrl}/${bundleName}.js` });
      }

       if (bundleName === APPLICATION_BUNDLE_NAME.RESIDENT) {
        const baseUrl = `${window.location.origin}${proxyBasePath}`;

        return LoaderUtils.loadScript({ src: `${baseUrl}/${bundleName}.js` });
      }

      return LoaderUtils.loadScript({
        src: `${remoteOrigin}/${bundleName}.js`,
        params: { type: "module" },
      });
    };

    const run = async () => {
      try {
        if (cancelled) {
          return;
        }

        await loadBundleScript();

        if (cancelled) {
          return;
        }

        const app = getEmbeddedApp(window, bundleName as ApplicationBundleName);

        const basePathProps = basePath ? { basePath: basePath } : {};

        const mergedConfig: IAppConfig = {
          env: "qa",
          featureFlags: {},
          hostType: "react",
          ...basePathProps,
          ...config,
        };

        app?.initialize(el, mergedConfig);
      } catch (e) {
        console.error("[MicrofrontendHost] load failed", e);
      }
    };

    void run();

    return () => {
      cancelled = true;

      const app = getEmbeddedApp(window, bundleName as ApplicationBundleName);

      app?.destroy();
    };
  }, [hostEl, bundleName, cssBundleName, proxyBasePath, remoteOrigin, config, basePath]);

  return createElement("app-external", {
    ref: setHostEl,
    style: { height: "100%", flexDirection: "column" },
  });
}
