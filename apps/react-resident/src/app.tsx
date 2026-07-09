import '@mantine/core/styles.css';

import { QueryProvider } from './app/providers/query-provider';
import { ThemeProvider } from './app/providers/theme-provider';
import { ResidentPage } from './pages/resident';

export function App() {
  return (
    <QueryProvider>
      <ThemeProvider>
        <ResidentPage />
      </ThemeProvider>
    </QueryProvider>
  );
}
