export const taskKeys = {
  all: ['tasks'] as const,
  detail: (id: number) => ['tasks', id] as const,
}
