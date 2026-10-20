# vue-torrent-bot

Vite IIFE remote micro-frontend (Vue).

- **Port:** 8084
- **Stack:** Vue 3, TypeScript, Vite, TanStack Query (`@tanstack/vue-query`), `@mantine-vue/core`
- **Structure:** FSD-lite — `app/` (providers) → `pages/<screen>/{ui,model}` → `entities/task` → `shared/{api,ui}`
- **Bundler:** Vite (library mode, IIFE)
- **Role:** Remote MF — exposes `window["ext-apps"]` (implements `IAppComponent`)
- **Host:** `react-entrypoint` page `/vue-torrent-bot`

## Commands

| Command | Description |
|---------|-------------|
| `nx serve vue-torrent-bot` | Build + serve the bundle on port 8084 (what the host loads) |
| `nx build vue-torrent-bot` | Production build → `dist/vue-torrent-bot.js` |
| `nx type-check vue-torrent-bot` | `vue-tsc` |
| `nx storybook vue-torrent-bot` | Storybook component workshop on port 6006 |
| `npm run build-storybook` | Static Storybook build → `storybook-static/` |
| `npm run dev` | Standalone Vite dev server (HMR; not the MFE entry) |

## Storybook

Write shared components here first, check them in Storybook, then use them in pages.

- Config: `.storybook/main.ts` (`defineMain`, framework `@storybook/vue3-vite`, stories glob
  `src/**/*.stories.@(js|ts)`). It strips the MFE-only Vite plugins and `build.lib` from
  `vite.config.ts`, because Storybook builds its own app instead of the IIFE bundle.
- `.storybook/preview.ts` registers `VueQueryPlugin` and wraps every story in `Providers.vue`
  (`MantineProvider`), plus imports `@mantine-vue/core/styles.css` — so components get the same
  Mantine/Query context as in the app.
- A story lives next to its component: `src/shared/ui/hero-card/hero-card.stories.ts` with
  `title: 'shared/ui/HeroCard'`. Copy that file as the template.
- Stories are dev-only: the MFE bundle is built from `src/remote-entry.ts` only, and
  `storybook-static/` is gitignored.

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
