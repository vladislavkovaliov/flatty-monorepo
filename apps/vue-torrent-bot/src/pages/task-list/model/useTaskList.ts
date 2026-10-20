import { useQuery } from '@tanstack/vue-query'
import { taskKeys } from '@/entities/task'
import { listTasks } from '@/shared/api/tasks'

export function useTaskList() {
  return useQuery({
    queryKey: taskKeys.all,
    queryFn: listTasks,
  })
}
