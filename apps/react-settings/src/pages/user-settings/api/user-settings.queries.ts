import { queryOptions, useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { apiClient } from '../../../shared/api/client';
import { SETTINGS_STORAGE_KEY } from './user-settings.mocks';
import type { UserSettings } from '../model/types';

export const USER_SETTINGS_QUERIES = {
  all: () => ['user-settings'] as const,
  current: () =>
    queryOptions({
      queryKey: [...USER_SETTINGS_QUERIES.all(), 'current'],
      queryFn: () => apiClient.get<UserSettings>(SETTINGS_STORAGE_KEY),
    }),
};

export function useUserSettings() {
  return useQuery(USER_SETTINGS_QUERIES.current());
}

export function useUpdateUserSettings() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (patch: Partial<UserSettings>) => {
      const currentRaw = localStorage.getItem(SETTINGS_STORAGE_KEY);
      const current: UserSettings = currentRaw ? JSON.parse(currentRaw) : {};
      const next: UserSettings = { ...current, ...patch };
      return apiClient.set(SETTINGS_STORAGE_KEY, next);
    },
    onMutate: async (patch) => {
      await queryClient.cancelQueries({ queryKey: USER_SETTINGS_QUERIES.current().queryKey });
      const prev = queryClient.getQueryData<UserSettings | null>(
        USER_SETTINGS_QUERIES.current().queryKey,
      );
      queryClient.setQueryData<UserSettings>(USER_SETTINGS_QUERIES.current().queryKey, (old) => {
        if (!old) return patch as UserSettings;
        return { ...old, ...patch };
      });
      return { prev };
    },
    onError: (_err, _patch, context) => {
      if (context?.prev) {
        queryClient.setQueryData(USER_SETTINGS_QUERIES.current().queryKey, context.prev);
      }
    },
    onSettled: () => {
      queryClient.invalidateQueries({ queryKey: USER_SETTINGS_QUERIES.current().queryKey });
    },
  });
}
