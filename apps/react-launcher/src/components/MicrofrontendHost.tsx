import { useEffect, useRef, type FC } from 'react';
import { loadBundle } from './LoaderUtils';
import type { IAppComponent } from '@flatty-budget/shared';

interface MicrofrontendHostProps {
  bundleName: string;
  cssBundleName?: string;
  remoteOrigin: string;
  proxyBasePath: string;
}

export const MicrofrontendHost: FC<MicrofrontendHostProps> = ({
  bundleName,
  proxyBasePath,
}) => {
  const hostRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    const host = hostRef.current;
    if (!host) return;

    let app: IAppComponent | undefined;

    loadBundle('ext-apps', bundleName, proxyBasePath)
      .then((module) => {
        app = module;
        app.initialize(host, {});
      })
      .catch(console.error);

    return () => {
      app?.destroy();
    };
  }, [bundleName, proxyBasePath]);

  return <div ref={hostRef} style={{ border: '1px solid #ccc', padding: 16 }} />;
};
