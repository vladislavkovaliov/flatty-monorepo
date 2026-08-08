"use client";

import {
  type ApplicationConfig,
  useApplications,
} from "@flatty-budget/mfe-host";
import { Loader } from "@mantine/core";
import dynamic from "next/dynamic";

const MicrofrontendHost = dynamic(
  () => import("@/core/micro-frontend-host").then((m) => m.MicrofrontendHost),
  {
    ssr: false,
    loading: () => <Loader />,
  },
);

// Fallback when the registry fetch errors or the row is missing.
const LEGACY_APP = {
  name: "settings",
  bundleJs: "settings",
  styleUrl: "styles",
  remoteOrigin: "http://localhost:8081",
  proxyBasePath: "/external-settings",
  basePath: "/",
} satisfies ApplicationConfig;

export default function SettingsPage() {
  const { data, loading } = useApplications();

  if (loading) {
    return <Loader />;
  }

  // On fetch error `data` is null → fall back to the legacy hardcoded config.
  // Each optional field falls back per-field too (partial rows stay resilient).
  const app = data?.find((row) => row.name === "settings") ?? LEGACY_APP;

  return (
    <MicrofrontendHost
      bundleName={app.bundleJs ?? LEGACY_APP.bundleJs}
      cssBundleName={app.styleUrl ?? LEGACY_APP.styleUrl}
      proxyBasePath={app.proxyBasePath ?? LEGACY_APP.proxyBasePath}
      basePath={app.basePath ?? LEGACY_APP.basePath}
      remoteOrigin={app.remoteOrigin ?? LEGACY_APP.remoteOrigin}
      config={{ env: "qa", featureFlags: {}, hostType: "react" }}
    />
  );
}
