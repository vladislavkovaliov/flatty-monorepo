import { queryOptions, useQuery } from '@tanstack/react-query';
import { getJson } from '../lib/http';
import type {
    UserListData,
    UserDetailData,
} from '../types/api';

export const USERS_QUERIES = {
    all: () => ['users'] as const,
    list: (limit: number, offset: number) =>
        queryOptions({
            queryKey: [...USERS_QUERIES.all(), 'list', { limit, offset }],
            queryFn: () => getJson<UserListData>(`/api/user?limit=${limit}&offset=${offset}`),
        }),
    byId: (id: string) =>
        queryOptions({
            queryKey: [...USERS_QUERIES.all(), 'byId', id],
            queryFn: () => getJson<UserDetailData>(`/api/user/${id}`),
        }),
};

export function useUsers(limit = 10, offset = 0) {
    return useQuery(USERS_QUERIES.list(limit, offset));
}

export function useUser(id: string) {
    return useQuery(USERS_QUERIES.byId(id));
}
