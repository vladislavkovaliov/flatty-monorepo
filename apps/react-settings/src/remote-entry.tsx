import { createRoot, type Root } from 'react-dom/client';
import { App } from './app';
import type { IAppComponent, IAppConfig } from '@flatty-budget/mfe-host';

const roots = new Map<HTMLElement, Root>();

const app: IAppComponent = {
  initialize(element: HTMLElement, _config: IAppConfig): void {
    const root = createRoot(element);
    root.render(<App />);
    roots.set(element, root);
  },
  destroy(): void {
    for (const root of roots.values()) {
      root.unmount();
    }
    roots.clear();
  },
};

export const entry = app;
