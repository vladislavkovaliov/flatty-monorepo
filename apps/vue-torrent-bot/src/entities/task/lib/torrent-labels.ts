import type { MediaType, Quality } from '../model/types'

export const MEDIA_TYPE_LABELS: Record<MediaType, string> = {
  movie: 'Фильм',
  series: 'Сериал',
}

export const QUALITY_LABELS: Record<Quality, string> = {
  '2160p': '2160p',
  '1080p': '1080p',
  '720p': '720p',
  unknown: 'Не определять',
}

// Object key order keeps the option order users see in the bot.
export const MEDIA_TYPE_OPTIONS = (Object.keys(MEDIA_TYPE_LABELS) as MediaType[]).map((value) => ({
  value,
  label: MEDIA_TYPE_LABELS[value],
}))

export const QUALITY_OPTIONS = (Object.keys(QUALITY_LABELS) as Quality[]).map((value) => ({
  value,
  label: QUALITY_LABELS[value],
}))

export function savePath(mediaType: MediaType): string {
  return mediaType === 'movie' ? 'Movies/' : 'TV/'
}
