import { Select, Stack, Text } from '@mantine/core';
import { useUserSettings, useUpdateUserSettings } from '../api/user-settings.queries';
import { DATE_FORMAT_OPTIONS, TIMEZONE_OPTIONS } from '../api/user-settings.mocks';
import type { DateFormat, TimezoneOffset } from '../model/types';

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
  const { data: settings } = useUserSettings();
  const { mutate } = useUpdateUserSettings();

  const dateFormat = settings?.dateFormat ?? getDefaultDateFormat();
  const timezone = settings?.timezone ?? getDefaultTimezone();

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
            mutate({ dateFormat: v as DateFormat });
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
            mutate({ timezone: v as TimezoneOffset });
          }
        }}
        searchable
        clearable={false}
      />
    </Stack>
  );
}
