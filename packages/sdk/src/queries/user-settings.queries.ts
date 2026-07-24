import { queryOptions, useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { getJson, putJson } from '../lib/http';

import type { 
    DtoUserSettingsResponse, 
    DtoUpdateUserSettingsRequest,
} from "../types/api"

export const USER_SETTINGS_QUERIES = {
  all: () => ['user-settings'] as const,
  current: () =>
    queryOptions({
      queryKey: [...USER_SETTINGS_QUERIES.all(), 'current'],
      queryFn: () => getJson<DtoUserSettingsResponse>('/api/user/me/settings'),
    }),
};

export function useUserSettings() {
    return useQuery(USER_SETTINGS_QUERIES.current());
}

export interface UserSettingsForm {
    dateFormat?: string;
    language?: string;
    theme?: string;
    timezone?: string;
}

export function toApiBody(data: UserSettingsForm): DtoUpdateUserSettingsRequest {
    return {
        date_format: data.dateFormat, 
        language: data.language,
        theme: data.theme,
        timezone: data.timezone
    };
}

export function useUpdateUserSettings() {
    const queryClient = useQueryClient();

    return useMutation({
        mutationFn: ({ data }: { data: UserSettingsForm }) => {
            return putJson<DtoUpdateUserSettingsRequest>('/api/user/me/settings', toApiBody(data))
        },
        onSettled: () => queryClient.invalidateQueries({ queryKey: USER_SETTINGS_QUERIES.all() }),
    });
}