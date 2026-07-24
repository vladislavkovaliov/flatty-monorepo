import { Select, Stack, Text } from '@mantine/core';
import type { Theme } from '../model/types';
import { useUserSettingsContext } from '../../../app/providers/user-settings-provider';

export const THEME_OPTIONS = [
  { value: 'light', label: 'Light' },
  { value: 'dark', label: 'Dark' },
  { value: 'system', label: 'System' },
] as const;

function getSystemTheme(): Theme {
  if (window.matchMedia('(prefers-color-scheme: dark)').matches) return 'dark';
  return 'light';
}

export function ThemesTab() {
  const userSettings = useUserSettingsContext();    
  const value = userSettings.settings?.theme ?? getSystemTheme();

  return (
    <Stack gap="md">
      <Text size="sm" c="dimmed">
        Choose your preferred appearance. If not set, your system preference is used.
      </Text>
      <Select
        label="Theme"
        placeholder="Pick a theme"
        data={[...THEME_OPTIONS]}
        value={value}
        onChange={(v) => {
          if (v) {
            userSettings.updateTheme(v);
          }
        }}
        clearable={false}
      />
    </Stack>
  );
}
