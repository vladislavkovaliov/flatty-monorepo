import { defineConfig } from 'vite';
import react from '@vitejs/plugin-react';

export default defineConfig({
  plugins: [react()],
  server: {
    port: 5175,
    proxy: {
      '/external-app': {
        target: 'http://localhost:8080',
        changeOrigin: true,
        rewrite: (p) => p.replace(/^\/external-app/, ''),
      },
      '/external-settings': {
        target: 'http://localhost:5174',
        changeOrigin: true,
        rewrite: (p) => p.replace(/^\/external-settings/, ''),
      },
    },
  },
});
