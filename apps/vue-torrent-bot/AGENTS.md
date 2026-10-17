# vue-torrent-bot

Vite IIFE remote micro-frontend (Vue).

- **Port:** 8084
- **Stack:** Vue 3, TypeScript, Vite, Pinia, `@mantine-vue/core`
- **Bundler:** Vite (library mode, IIFE)
- **Role:** Remote MF — exposes `window["ext-apps"]` (implements `IAppComponent`)
- **Host:** `react-entrypoint` page `/vue-torrent-bot`

## Commands

| Command | Description |
|---------|-------------|
| `nx serve vue-torrent-bot` | Build + serve the bundle on port 8084 (what the host loads) |
| `nx build vue-torrent-bot` | Production build → `dist/vue-torrent-bot.js` |
| `nx type-check vue-torrent-bot` | `vue-tsc` |
| `npm run dev` | Standalone Vite dev server (HMR; not the MFE entry) |

## Micro-frontend notes

- Entry: `src/remote-entry.ts` — mounts `MantineProvider` + `App.vue` into the host element and assigns
  `window["ext-apps"] = remote`.
- Bundle: IIFE at `dist/vue-torrent-bot.js`; CSS is injected by JS (`vite-plugin-css-injected-by-js`)
  because the host loads only the JS file.
- `define: { "process.env.NODE_ENV": "production" }` in `vite.config.ts` is required — Vite lib mode does
  not replace it, and the raw `<script>` would crash on `process is not defined`.
- Registry: `applications` row `vue-torrent-bot` — migration
  `apps/go-api/migrations/008_add_vue_torrent_bot_application.sql` (`<name>.js` served via
  `/external-vue-torrent-bot` rewrite).
- Rebuild after changes; the host caches the script (hard reload).
