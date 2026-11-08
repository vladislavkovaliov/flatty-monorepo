import { useQuery } from '@tanstack/vue-query'
import { taskKeys } from '@/entities/task'
import { listDownloadingTasks } from '@/shared/api/tasks'

export function useDownloadingTasks() {
  return useQuery({
    queryKey: taskKeys.downloadingCount,
    queryFn: listDownloadingTasks,
  })
}
