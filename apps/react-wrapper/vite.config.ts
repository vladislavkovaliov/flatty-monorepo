import path from "node:path";

import { defineConfig } from "vite";
import react, { reactCompilerPreset } from "@vitejs/plugin-react";
import babel from "@rolldown/plugin-babel";

// https://vite.dev/config/
export default defineConfig({
  resolve: {
    alias: {
      "#": path.resolve(__dirname, "src"),
    },
  },
  plugins: [
    react(),
    babel({ presets: [reactCompilerPreset()] }),
  ],
  server: {
    port: 5174,
    proxy: {
      // Same idea as apps/angular-wrapper/src/proxy.conf.json → static host.
      // react-app webpack dev-server is served on :8080 in the current setup.
      "/external-app": {
        target: "http://localhost:8080",
        changeOrigin: true,
        rewrite: (path) => path.replace(/^\/external-app/, ""),
      },
      "/external-settings": {
        target: "http://localhost:8081",
        changeOrigin: true,
        rewrite: (path) => path.replace(/^\/external-settings/, ""),
      },
      "/external-resident": {
        target: "http://localhost:8082",
        changeOrigin: true,
        rewrite: (path) => path.replace(/^\/external-resident/, ""),
      },
      "/api": { target: "http://localhost:8080" },
      "/graphql": { target: "http://localhost:3000" },
    },
  },
});
