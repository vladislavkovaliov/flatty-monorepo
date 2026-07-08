import '@mantine/core/styles.css';

import { QueryProvider } from './app/providers/query-provider';
import { ThemeProvider } from './app/providers/theme-provider';
import { UserSettingsPage } from './pages/user-settings';

export function App() {
  return (
    <QueryProvider>
      <ThemeProvider>
        <UserSettingsPage />
      </ThemeProvider>
    </QueryProvider>
  );
}
