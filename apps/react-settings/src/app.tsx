import '@mantine/core/styles.css';

import { QueryProvider } from './app/providers/query-provider';
import { ThemeProvider } from './app/providers/theme-provider';
import { UserSettingsPage } from './pages/user-settings';
import { UserSettingsProvider } from './app/providers/user-settings-provider';

export function App() {
  return (
    <QueryProvider>
      <ThemeProvider>
        <UserSettingsProvider>
          <UserSettingsPage />
        </UserSettingsProvider>
      </ThemeProvider>
    </QueryProvider>
  );
}
