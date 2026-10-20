// Форма совпадает с ответом python-torrent-bot (snake_case, FastAPI).

export type TorrentStatus =
  'queued' | 'downloading' | 'paused' | 'completed' | 'failed' | 'cancelled'

export type MediaType = 'movie' | 'series'

export type Quality = '2160p' | '1080p' | '720p' | 'unknown'

export interface TorrentTask {
  id: number
  user_id: number
  magnet: string
  info_hash: string
  name: string
  media_type: MediaType
  quality: Quality
  status: TorrentStatus
  progress: number
  size_bytes: number | null
  download_speed: number
  upload_speed: number
  eta_seconds: number | null
  created_at: string
  completed_at: string | null
  provider_id: string
}

export interface AddTaskInput {
  magnet: string
  media_type: MediaType
  quality: Quality
}
