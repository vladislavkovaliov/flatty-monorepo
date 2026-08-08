"use client";

import { Loader } from "@mantine/core";
import {
  useApplications,
  type ApplicationConfig,
} from "@flatty-budget/mfe-host";
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
  name: "resident",
  bundleJs: "resident",
  styleUrl: "styles",
  remoteOrigin: "http://localhost:8082",
  proxyBasePath: "/external-resident",
  basePath: "/",
} satisfies ApplicationConfig;

export default function ResidentPage() {
  const { data, loading } = useApplications();
  console.log({data})
  if (loading) {
    return <Loader />;
  }

  // On fetch error `data` is null → fall back to the legacy hardcoded config.
  // Each optional field falls back per-field too (partial rows stay resilient).
  const app = data?.find((row) => row.name === "resident") ?? LEGACY_APP;

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
