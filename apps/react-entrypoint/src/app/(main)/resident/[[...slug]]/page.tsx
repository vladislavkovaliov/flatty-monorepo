"use client";

import { Loader } from "@mantine/core";
import dynamic from "next/dynamic";

const MicrofrontendHost = dynamic(
  () => import("@/core/micro-frontend-host").then((m) => m.MicrofrontendHost),
  {
    ssr: false,
    loading: () => <Loader />,
  },
);

export default function ResidentPage() {
  return (
    <MicrofrontendHost
      bundleName="resident"
      cssBundleName="styles"
      proxyBasePath="/external-resident"
      basePath="/"
      config={{ env: "qa", featureFlags: {}, hostType: "react" }}
    />
  );
}
