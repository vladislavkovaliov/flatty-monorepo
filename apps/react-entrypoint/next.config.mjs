/** @type {import('next').NextConfig} */

const GO_API_ORIGIN = "http://localhost:8080";

// Fallback entries when the registry fetch fails at boot (go-api down, DB down, empty registry).
const HARDCODED_APP_REWRITES = [
  {
    source: "/external-settings/:path*",
    destination: "http://localhost:8081/:path*",
  },
  {
    source: "/external-resident/:path*",
    destination: "http://localhost:8082/:path*",
  },
  {
    source: "/external-app/:path*",
    destination: "http://localhost:8080/:path*",
  },
];

let registryPromise;

// Fetched once at boot; memoized for the process lifetime (gap-fill #8).
function fetchApplications() {
  if (!registryPromise) {
    registryPromise = fetch(`${GO_API_ORIGIN}/api/applications`)
      .then((res) => {
        if (!res.ok) {
          throw new Error(`HTTP ${res.status}`);
        }
        return res.json();
      })
      .then((body) => body.data ?? [])
      .catch((err) => {
        console.warn(
          "[next.config] Failed to fetch application registry from go-api; falling back to hardcoded rewrites.",
          err,
        );
        return null;
      });
  }
  return registryPromise;
}

const nextConfig = {
  reactCompiler: true,
  async rewrites() {
    const base = [
      {
        source: "/api/auth/:path*",
        destination: "http://localhost:3000/api/auth/:path*",
      },
      {
        source: "/graphql",
        destination: "http://localhost:3000/graphql",
      },
      {
        source: "/api/uploads",
        destination: "/api/uploads",
      },
      {
        source: "/api/:path*",
        destination: `${GO_API_ORIGIN}/api/:path*`,
      },
      // {
      //   source: "/admin/applications",
      //   destination: `http://localhost:8080/admin/applications`
      // }
    ];

    const apps = await fetchApplications();

    // Fall back on failure OR empty registry — otherwise a healthy-but-empty DB
    // would silently break /external-* proxying in dev.
    if (!apps || apps.length === 0) {
      return [...base, ...HARDCODED_APP_REWRITES];
    }

    const appRewrites = apps.map((app) => ({
      source: `${app.proxy_base_path}/:path*`,
      destination: `${app.remote_origin}/:path*`,
    }));

    return [...base, ...appRewrites];
  },
};

export default nextConfig;
