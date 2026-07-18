import { Select, Stack, Text } from '@mantine/core';
import { useUserSettings, useUpdateUserSettings } from '../api/user-settings.queries';
import { THEME_OPTIONS } from '../api/user-settings.mocks';
import type { Theme } from '../model/types';

function getSystemTheme(): Theme {
  if (window.matchMedia('(prefers-color-scheme: dark)').matches) return 'dark';
  return 'light';
}

export function ThemesTab() {
  const { data: settings } = useUserSettings();
  const { mutate } = useUpdateUserSettings();

  const value = settings?.theme ?? getSystemTheme();

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
            mutate({ theme: v as Theme });
          }
        }}
        clearable={false}
      />
    </Stack>
  );
}
