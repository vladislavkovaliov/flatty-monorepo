import { createApp, h, type App } from 'vue';
import { MantineProvider } from '@mantine-vue/core';
import '@mantine-vue/core/styles.css';
import Root from './App.vue';
import type { IAppComponent, IAppConfig } from '@flatty-budget/mfe-host';

let app: App | null = null;

const remote: IAppComponent = {
  initialize(element, config: IAppConfig) {
    app = createApp({
      render: () =>
        h(MantineProvider, { defaultColorScheme: 'auto' }, () => h(Root, { config })),
    });
    app.mount(element);
  },
  destroy() {
    app?.unmount();
    app = null;
  },
};

(window as any)['ext-apps'] = remote;

export default remote;
