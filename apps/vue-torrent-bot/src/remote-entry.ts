import { createApp, h, type App } from 'vue';
import { VueQueryPlugin } from '@tanstack/vue-query';
import '@mantine-vue/core/styles.css';
import Root from './app/App.vue';
import Providers from './app/providers/Providers.vue';
import { queryClient } from './app/providers/query';
import type { IAppComponent, IAppConfig } from '@flatty-budget/mfe-host';

let app: App | null = null;

const remote: IAppComponent = {
  initialize(element, config: IAppConfig) {
    app = createApp({
      render: () => h(Providers, null, () => h(Root, { config })),
    });
    app.use(VueQueryPlugin, { queryClient });
    app.mount(element);
  },
  destroy() {
    app?.unmount();
    app = null;
  },
};

(window as any)['ext-apps'] = remote;

export default remote;
