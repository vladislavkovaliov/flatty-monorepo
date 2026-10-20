import { setup } from '@storybook/vue3-vite'
import type { Preview } from '@storybook/vue3-vite'
import { VueQueryPlugin } from '@tanstack/vue-query'
import '@mantine-vue/core/styles.css'

import Providers from '../src/app/providers/Providers.vue'
import { queryClient } from '../src/app/providers/query'

// Same plugin stack as main.ts / remote-entry.ts, so stories behave like the real app.
setup((app) => {
  app.use(VueQueryPlugin, { queryClient })
})

const preview: Preview = {
  decorators: [
    () => ({
      components: { Providers },
      template: '<Providers><story /></Providers>',
    }),
  ],
}

export default preview
