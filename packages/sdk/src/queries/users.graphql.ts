import { queryOptions, useQuery } from '@tanstack/react-query';
import type {
    ListUserResponse,
    User,
} from '../types/graphql';

async function graphqlRequest<T>(query: string, variables: Record<string, unknown>): Promise<T> {
    const response = await fetch('/graphql', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        credentials: 'include',
        body: JSON.stringify({ query, variables }),
    });

    const json = await response.json();

    if (json.errors) {
        throw new Error(json.errors[0]?.message ?? 'GraphQL error');
    }

    return json.data as T;
}

const LIST_USERS = `
query userList($limit: Int, $offset: Int) {
    userList(limit: $limit, offset: $offset) {
        data {
            id
            name
            email
            emailVerified
            image
            createdAt
            updatedAt
        }
        total
    }
}
`;

const GET_USER = `
query user($id: String!) {
    user(id: $id) {
        id
        name
        email
        emailVerified
        image
        createdAt
        updatedAt
    }
    total
}
`;

type GqlListData = { userList: ListUserResponse };
type GqlGetData = { user: User };

export const USERS_GRAPHQL_QUERIES = {
    all: () => ['users', 'graphql'] as const,
    list: (limit = 10, offset = 0) =>
        queryOptions({
            queryKey: [...USERS_GRAPHQL_QUERIES.all(), 'list', { limit, offset }],
            queryFn: () => graphqlRequest<GqlListData>(LIST_USERS, { limit, offset }),
        }),
    byId: (id: string) =>
        queryOptions({
            queryKey: [...USERS_GRAPHQL_QUERIES.all(), 'byId', id],
            queryFn: () => graphqlRequest<GqlGetData>(GET_USER, { id }),
        }),
};

export function useUsersGraphql(limit = 10, offset = 0) {
    return useQuery(USERS_GRAPHQL_QUERIES.list(limit, offset));
}

export function useUserGraphql(id: string) {
    return useQuery(USERS_GRAPHQL_QUERIES.byId(id));
}
