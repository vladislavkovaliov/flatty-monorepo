import { Select, Stack, Text } from '@mantine/core';
import type { DateFormat, TimezoneOffset } from '../model/types';
import { useUserSettingsContext } from '../../../app/providers/user-settings-provider';

export const DATE_FORMAT_OPTIONS = [
  { value: 'DD/MM/YYYY', label: 'DD/MM/YYYY' },
  { value: 'MM/DD/YYYY', label: 'MM/DD/YYYY' },
  { value: 'YYYY-MM-DD', label: 'YYYY-MM-DD' },
] as const;

export const TIMEZONE_OPTIONS = [
  { value: 'UTC-12', label: 'UTC-12' },
  { value: 'UTC-11', label: 'UTC-11' },
  { value: 'UTC-10', label: 'UTC-10' },
  { value: 'UTC-9', label: 'UTC-9' },
  { value: 'UTC-8', label: 'UTC-8' },
  { value: 'UTC-7', label: 'UTC-7' },
  { value: 'UTC-6', label: 'UTC-6' },
  { value: 'UTC-5', label: 'UTC-5' },
  { value: 'UTC-4', label: 'UTC-4' },
  { value: 'UTC-3', label: 'UTC-3' },
  { value: 'UTC-2', label: 'UTC-2' },
  { value: 'UTC-1', label: 'UTC-1' },
  { value: 'UTC+0', label: 'UTC+0' },
  { value: 'UTC+1', label: 'UTC+1' },
  { value: 'UTC+2', label: 'UTC+2' },
  { value: 'UTC+3', label: 'UTC+3' },
  { value: 'UTC+4', label: 'UTC+4' },
  { value: 'UTC+5', label: 'UTC+5' },
  { value: 'UTC+6', label: 'UTC+6' },
  { value: 'UTC+7', label: 'UTC+7' },
  { value: 'UTC+8', label: 'UTC+8' },
  { value: 'UTC+9', label: 'UTC+9' },
  { value: 'UTC+10', label: 'UTC+10' },
  { value: 'UTC+11', label: 'UTC+11' },
  { value: 'UTC+12', label: 'UTC+12' },
] as const;

function getDefaultDateFormat(): DateFormat {
  const raw = new Intl.DateTimeFormat().resolvedOptions().locale;
  if (raw.startsWith('en-US')) return 'MM/DD/YYYY';
  if (raw.startsWith('ja') || raw.startsWith('zh') || raw.startsWith('ko')) return 'YYYY-MM-DD';
  return 'DD/MM/YYYY';
}

function getDefaultTimezone(): TimezoneOffset {
  const offset = -new Date().getTimezoneOffset();
  const hours = Math.floor(Math.abs(offset) / 60);
  const sign = offset >= 0 ? '+' : '-';
  const value = `UTC${sign}${hours}` as TimezoneOffset;
  const valid = TIMEZONE_OPTIONS.map((o) => o.value);
  return valid.includes(value) ? value : 'UTC+0';
}

export function TimezoneTab() {
  const userSettings = useUserSettingsContext();    
  
  const dateFormat = userSettings.settings?.dateFormat ?? getDefaultDateFormat();
  const timezone = userSettings.settings?.timezone ?? getDefaultTimezone();

  return (
    <Stack gap="md">
      <Text size="sm" c="dimmed">
        Configure date format and timezone. If not set, your system locale and timezone are used.
      </Text>
      <Select
        label="Date format"
        placeholder="Pick a date format"
        data={[...DATE_FORMAT_OPTIONS]}
        value={dateFormat}
        onChange={(v) => {
          if (v) {
            userSettings.updateTimezone(v, timezone);
          }
        }}
        clearable={false}
      />
      <Select
        label="Timezone"
        placeholder="Pick a timezone"
        data={[...TIMEZONE_OPTIONS]}
        value={timezone}
        onChange={(v) => {
          if (v) {
            userSettings.updateTimezone(dateFormat, v);
          }
        }}
        searchable
        clearable={false}
      />
    </Stack>
  );
}
