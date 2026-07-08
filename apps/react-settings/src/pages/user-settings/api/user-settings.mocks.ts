export const LOCALE_OPTIONS = [
  { value: 'en', label: 'English' },
  { value: 'ru', label: 'Русский' },
  { value: 'de', label: 'Deutsch' },
  { value: 'fr', label: 'Français' },
  { value: 'es', label: 'Español' },
  { value: 'zh', label: '中文' },
  { value: 'ja', label: '日本語' },
] as const;

export const THEME_OPTIONS = [
  { value: 'light', label: 'Light' },
  { value: 'dark', label: 'Dark' },
] as const;

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

export const SETTINGS_STORAGE_KEY = 'flatty-budget:user-settings';
