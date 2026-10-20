import { useQuery } from '@tanstack/vue-query'

export interface BotSettings {
  download_path: string
  movies_path: string
  tv_path: string
}

export const settingsQueryKey = ['settings'] as const

// Заглушка: заменить на реальный запрос, когда у python-torrent-bot появится /api/settings
export function useSettings() {
  return useQuery({
    queryKey: settingsQueryKey,
    queryFn: async (): Promise<BotSettings> => ({
      download_path: '/mnt/dns320/Downloads',
      movies_path: '/mnt/dns320/Movies',
      tv_path: '/mnt/dns320/TV',
    }),
  })
}
