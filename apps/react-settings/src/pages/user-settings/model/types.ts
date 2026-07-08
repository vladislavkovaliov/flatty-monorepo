export type Locale = 'en' | 'ru' | 'de' | 'fr' | 'es' | 'zh' | 'ja';

export type Theme = 'light' | 'dark';

export type DateFormat = 'DD/MM/YYYY' | 'MM/DD/YYYY' | 'YYYY-MM-DD';

export type TimezoneOffset =
  | 'UTC-12' | 'UTC-11' | 'UTC-10' | 'UTC-9' | 'UTC-8' | 'UTC-7'
  | 'UTC-6' | 'UTC-5' | 'UTC-4' | 'UTC-3' | 'UTC-2' | 'UTC-1'
  | 'UTC+0'
  | 'UTC+1' | 'UTC+2' | 'UTC+3' | 'UTC+4' | 'UTC+5' | 'UTC+6'
  | 'UTC+7' | 'UTC+8' | 'UTC+9' | 'UTC+10' | 'UTC+11' | 'UTC+12';

export interface UserSettings {
  locale: Locale | null;
  theme: Theme | null;
  dateFormat: DateFormat | null;
  timezone: TimezoneOffset | null;
}
