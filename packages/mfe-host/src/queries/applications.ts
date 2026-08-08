import { useEffect, useState } from "react";

export interface ApplicationConfig {
  name: string;
  bundleJs?: string;
  styleUrl?: string;
  remoteOrigin?: string;
  proxyBasePath?: string;
  basePath?: string;
}

interface ApiApplicationRow {
  id: number;
  name: string;
  env: string;
  bundle_js: string;
  style_url?: string;
  remote_origin: string;
  proxy_base_path: string;
  base_path: string;
}

interface ApiApplicationsResponse {
  data: ApiApplicationRow[];
}

export interface UseApplicationsResult {
  data: ApplicationConfig[] | null;
  loading: boolean;
  error: Error | null;
}

function toApplicationConfig(row: ApiApplicationRow): ApplicationConfig {
  return {
    name: row.name,
    bundleJs: row.bundle_js,
    styleUrl: row.style_url,
    remoteOrigin: row.remote_origin,
    proxyBasePath: row.proxy_base_path,
    basePath: row.base_path,
  };
}

/**
 * Dependency-free registry hook: plain fetch + useState/useEffect.
 * Reads the public go-api registry (GET /api/applications, proxied same-origin).
 */
export function useApplications(): UseApplicationsResult {
  const [data, setData] = useState<ApplicationConfig[] | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<Error | null>(null);

  useEffect(() => {
    const controller = new AbortController();
    let cancelled = false;

    fetch("/api/applications", { signal: controller.signal })
      .then((res) => {
        if (!res.ok) {
          throw new Error(`Failed to fetch applications: HTTP ${res.status}`);
        }
        return res.json() as Promise<ApiApplicationsResponse>;
      })
      .then((body) => {
        if (cancelled) {
          return;
        }
        setData((body.data ?? []).map(toApplicationConfig));
        setLoading(false);
      })
      .catch((err: unknown) => {
        if (cancelled) {
          return;
        }
        setError(err instanceof Error ? err : new Error(String(err)));
        setLoading(false);
      });

    return () => {
      cancelled = true;
      controller.abort();
    };
  }, []);

  return { data, loading, error };
}
