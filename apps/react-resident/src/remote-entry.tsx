import { createRoot, type Root } from 'react-dom/client';
import { App } from './app';
import type { IAppComponent, IAppConfig } from '@flatty-budget/mfe-host';

const roots = new Map<HTMLElement, Root>();

const app: IAppComponent = {
  initialize(element: HTMLElement, _config: IAppConfig): void {

    console.log('init')
    const root = createRoot(element);
    root.render(<App />);
    roots.set(element, root);
  },
  destroy(): void {
    const elements = Array.from(roots.keys());
    for (const element of elements) {
      const root = roots.get(element);
      if (root) {
        root.unmount();
      }
      roots.delete(element);
    }
  },
};

export default app;
