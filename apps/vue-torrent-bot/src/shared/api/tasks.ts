import type { AddTaskInput, TorrentTask } from '@/entities/task'
import { apiFetch } from './client'

export function listTasks(): Promise<TorrentTask[]> {
  return apiFetch<TorrentTask[]>('/tasks')
}

export function listCompletedTasks(): Promise<Number> {
  return apiFetch<Number>('/tasks/completed')
}

export function listDownloadingTasks(): Promise<Number> {
  return apiFetch<Number>('/tasks/downloading')
}

export function listQueuedTasks(): Promise<Number> {
  return apiFetch<Number>('/tasks/queued')
}

export function getTask(id: number): Promise<TorrentTask> {
  return apiFetch<TorrentTask>(`/tasks/${id}`)
}

export function addTask(input: AddTaskInput): Promise<TorrentTask> {
  return apiFetch<TorrentTask>('/tasks', {
    method: 'POST',
    body: JSON.stringify(input),
  })
}

export function pauseTask(id: number): Promise<TorrentTask> {
  return apiFetch<TorrentTask>(`/tasks/${id}/pause`, { method: 'POST' })
}

export function resumeTask(id: number): Promise<TorrentTask> {
  return apiFetch<TorrentTask>(`/tasks/${id}/resume`, { method: 'POST' })
}

export function cancelTask(id: number): Promise<TorrentTask> {
  return apiFetch<TorrentTask>(`/tasks/${id}/cancel`, { method: 'POST' })
}

export function deleteTask(id: number): Promise<void> {
  return apiFetch<void>(`/tasks/${id}`, { method: 'DELETE' })
}
