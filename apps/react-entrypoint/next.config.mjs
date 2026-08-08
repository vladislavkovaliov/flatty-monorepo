/** @type {import('next').NextConfig} */
const nextConfig = {
  reactCompiler: true,
  async rewrites() {
    return [
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
        destination: "http://localhost:8080/api/:path*",
      },
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
  },
};

export default nextConfig;
