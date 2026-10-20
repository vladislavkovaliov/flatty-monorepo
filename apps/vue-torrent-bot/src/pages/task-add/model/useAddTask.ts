import { useMutation, useQueryClient } from '@tanstack/vue-query'
import { taskKeys } from '@/entities/task'
import { addTask } from '@/shared/api/tasks'

export function useAddTask() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: addTask,
    onSuccess: () => queryClient.invalidateQueries({ queryKey: taskKeys.all }),
  })
}
