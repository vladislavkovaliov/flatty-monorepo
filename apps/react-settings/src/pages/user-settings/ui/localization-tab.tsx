import { Select, Stack, Text } from '@mantine/core';
import type { Locale } from '../model/types';
import { useUserSettingsContext } from '../../../hooks/use-user-settings-context';
import { LOCALE_OPTIONS } from './constants';

function getDefaultLocale(): Locale {
  const raw = navigator.language?.slice(0, 2);
  const valid: Locale[] = ['en', 'ru', 'de', 'fr', 'es', 'zh', 'ja'];
  return valid.includes(raw as Locale) ? (raw as Locale) : 'en';
}

export function LocalizationTab() {
  const userSettings = useUserSettingsContext();    
  const value = userSettings.settings?.language ?? getDefaultLocale();

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
            userSettings.updateLanguage(v);
          }
        }}
        searchable
        clearable={false}
      />
    </Stack>
  );
}
