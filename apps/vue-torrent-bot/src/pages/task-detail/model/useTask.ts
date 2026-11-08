import { computed, toValue, type MaybeRefOrGetter } from 'vue'
import { useQuery } from '@tanstack/vue-query'
import { taskKeys } from '@/entities/task'
import { getTask } from '@/shared/api/tasks'

export function useTask(taskId: MaybeRefOrGetter<number>) {
  return useQuery({
    queryKey: computed(() => taskKeys.detail(toValue(taskId))),
    queryFn: () => getTask(toValue(taskId)),
  })
}
