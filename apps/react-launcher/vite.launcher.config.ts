import { defineConfig } from "vite";
import react, { reactCompilerPreset } from "@vitejs/plugin-react";
import babel from "@rolldown/plugin-babel";
import * as path from "node:path";

const entryPath = path.resolve(__dirname, "./src/Launcher.tsx");

export default defineConfig(({ mode }) => {
  const isDevelopment = mode === "development";
  const portArgIndex = process.argv.indexOf('--port');
  const port =
      portArgIndex !== -1 ? Number(process.argv[portArgIndex + 1]) : undefined;
  console.log(mode)
  return {
    base: "./",
    root: path.resolve(__dirname, "."),
    server: {
      port: port,
      proxy: {
        // "/external-app": {
        //   target: "http://localhost:8080",
        //   changeOrigin: true,
        //   rewrite: (path) => path.replace(/^\/external-app/, ""),
        // },
        "/external-settings": {
          target: "http://localhost:8081",
          changeOrigin: true,
          rewrite: (path) => path.replace(/^\/external-settings/, ""),
        },
      },
    },
    define: {
      "process.env.NODE_ENV": JSON.stringify(
        isDevelopment ? "development" : "production",
      ),
    },
    build: {
      lib: {
        entry: entryPath,
        name: "Launcher",
        formats: ["es", "umd"],
        fileName: () => "launcher.js",
      },
      outDir: path.resolve(__dirname, "./dist"),
    },
    plugins: [react(), babel({ presets: [reactCompilerPreset()] })],
  };

});