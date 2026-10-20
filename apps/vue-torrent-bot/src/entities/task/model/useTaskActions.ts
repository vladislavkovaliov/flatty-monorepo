import { useMutation, useQueryClient } from '@tanstack/vue-query'
import { cancelTask, deleteTask, pauseTask, resumeTask } from '@/shared/api/tasks'
import { taskKeys } from './queryKeys'

export function useTaskActions() {
  const queryClient = useQueryClient()
  const invalidate = () => queryClient.invalidateQueries({ queryKey: taskKeys.all })

  return {
    pause: useMutation({ mutationFn: pauseTask, onSuccess: invalidate }),
    resume: useMutation({ mutationFn: resumeTask, onSuccess: invalidate }),
    cancel: useMutation({ mutationFn: cancelTask, onSuccess: invalidate }),
    remove: useMutation({ mutationFn: deleteTask, onSuccess: invalidate }),
  }
}
