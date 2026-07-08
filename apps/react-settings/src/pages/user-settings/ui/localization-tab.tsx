import { Select, Stack, Text } from '@mantine/core';
import { useUserSettings, useUpdateUserSettings } from '../api/user-settings.queries';
import { LOCALE_OPTIONS } from '../api/user-settings.mocks';
import type { Locale } from '../model/types';

function getDefaultLocale(): Locale {
  const raw = navigator.language?.slice(0, 2);
  const valid: Locale[] = ['en', 'ru', 'de', 'fr', 'es', 'zh', 'ja'];
  return valid.includes(raw as Locale) ? (raw as Locale) : 'en';
}

export function LocalizationTab() {
  const { data: settings } = useUserSettings();
  const { mutate } = useUpdateUserSettings();

  const value = settings?.locale ?? getDefaultLocale();

  return (
    <Stack gap="md">
      <Text size="sm" c="dimmed">
        Select your preferred language. If not set, your browser language is used.
      </Text>
      <Select
        label="Language"
        placeholder="Pick a language"
        data={[...LOCALE_OPTIONS]}
        value={value}
        onChange={(v) => {
          if (v) {
            mutate({ locale: v as Locale });
            console.log('[UserSettings] locale updated:', v);
          }
        }}
        searchable
        clearable={false}
      />
    </Stack>
  );
}
