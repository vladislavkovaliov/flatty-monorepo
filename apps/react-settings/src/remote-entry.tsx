import { createRoot, type Root } from 'react-dom/client';
import { App } from './app';
import type { IAppComponent } from '@flatty-budget/shared';

const roots = new Map<HTMLElement, Root>();

const app: IAppComponent = {
  initialize(element: HTMLElement, _config: Record<string, unknown>): void {
    const root = createRoot(element);
    root.render(<App />);
    roots.set(element, root);
  },
  destroy(element: HTMLElement): void {
    const root = roots.get(element);
    if (root) {
      root.unmount();
      roots.delete(element);
    }
  },
};

export default app;
