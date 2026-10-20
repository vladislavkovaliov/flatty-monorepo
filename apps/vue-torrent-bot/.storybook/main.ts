import { defineMain } from '@storybook/vue3-vite/node'

// vite.config.ts is tuned for the MFE bundle (library/IIFE mode) — Storybook builds its own app.
const mfeOnlyPlugins = ['vite-plugin-css-injected-by-js', 'vite-plugin-vue-devtools']

export default defineMain({
  framework: { name: '@storybook/vue3-vite', options: {} },
  stories: ['../src/**/*.mdx', '../src/**/*.stories.@(js|ts)'],
  addons: ['@storybook/addon-docs'],
  viteFinal: async (config) => {
    if (config.build) {
      delete config.build.lib
      delete config.build.rollupOptions?.output
    }
    const plugins = (config.plugins ?? []).flat(Infinity) as { name?: string }[]
    config.plugins = plugins.filter(
      (plugin) => !mfeOnlyPlugins.includes(plugin.name ?? ''),
    ) as typeof config.plugins
    return config
  },
})
