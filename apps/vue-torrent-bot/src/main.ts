import { createApp, h } from 'vue'
import { VueQueryPlugin } from '@tanstack/vue-query'
import '@mantine-vue/core/styles.css'
import App from './app/App.vue'
import Providers from './app/providers/Providers.vue'
import { queryClient } from './app/providers/query'
import type { IAppConfig } from '@flatty-budget/mfe-host'

// Standalone-режим (`npm run dev`); в MFE используется remote-entry.ts.
const standaloneConfig: IAppConfig = {
  env: 'development',
  featureFlags: {},
  hostType: 'other',
}

const app = createApp({
  render: () => h(Providers, null, () => h(App, { config: standaloneConfig })),
})

app.use(VueQueryPlugin, { queryClient })
app.mount('#app')
