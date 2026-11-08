export const taskKeys = {
  all: ['tasks'] as const,
  downloadingCount: ['tasks-downloading-count'] as const,
  detail: (id: number) => ['tasks', id] as const,
}
