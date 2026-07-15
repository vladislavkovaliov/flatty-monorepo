import { queryOptions, useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import type {
  DeleteResidentLocationResponse,
  ListResidentLocationResponse,
  ResidentLocation,
  ResidentLocationInput,
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

const LIST_RESIDENT_LOCATION = `
  query ListResidentLocation($limit: Int, $offset: Int) {
    list(limit: $limit, offset: $offset) {
      data {
        id
        country
        city
        postalCode
        street
        house
        apartment
        createdAt
        updatedAt
      }
      total
    }
  }
`;

const CREATE_RESIDENT_LOCATION = `
  mutation CreateResidentLocation($input: ResidentLocationInput!) {
    create(residentLocatoinData: $input) {
      id
      country
      city
      postalCode
      street
      house
      apartment
      createdAt
      updatedAt
    }
  }
`;

const UPDATE_RESIDENT_LOCATION = `
  mutation UpdateResidentLocation($id: Int!, $input: ResidentLocationInput!) {
    update(id: $id, residentLocatoinData: $input) {
      id
      country
      city
      postalCode
      street
      house
      apartment
      createdAt
      updatedAt
    }
  }
`;

const DELETE_RESIDENT_LOCATION = `
  mutation DeleteResidentLocation($id: Int!) {
    delete(id: $id) {
      data
    }
  }
`;

type GqlListData = { list: ListResidentLocationResponse };
type GqlCreateData = { create: ResidentLocation };
type GqlUpdateData = { update: ResidentLocation };
type GqlDeleteData = { delete: DeleteResidentLocationResponse };

export const RESIDENT_LOCATION_GRAPHQL_QUERIES = {
  all: () => ['resident-location', 'graphql'] as const,
  list: (limit = 10, offset = 0) =>
    queryOptions({
      queryKey: [...RESIDENT_LOCATION_GRAPHQL_QUERIES.all(), 'list', { limit, offset }],
      queryFn: () => graphqlRequest<GqlListData>(LIST_RESIDENT_LOCATION, { limit, offset }),
    }),
};

export function useResidentLocationGraphql(limit = 10, offset = 0) {
  return useQuery(RESIDENT_LOCATION_GRAPHQL_QUERIES.list(limit, offset));
}

export function useCreateResidentLocationGraphql() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (input: ResidentLocationInput) =>
      graphqlRequest<GqlCreateData>(CREATE_RESIDENT_LOCATION, { input }),
    onSettled: () =>
      queryClient.invalidateQueries({ queryKey: RESIDENT_LOCATION_GRAPHQL_QUERIES.all() }),
  });
}

export function useUpdateResidentLocationGraphql() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: ({ id, data }: { id: number; data: ResidentLocationInput }) =>
      graphqlRequest<GqlUpdateData>(UPDATE_RESIDENT_LOCATION, { id, input: data }),
    onSettled: () =>
      queryClient.invalidateQueries({ queryKey: RESIDENT_LOCATION_GRAPHQL_QUERIES.all() }),
  });
}

export function useDeleteResidentLocationGraphql() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (id: number) =>
      graphqlRequest<GqlDeleteData>(DELETE_RESIDENT_LOCATION, { id }),
    onSettled: () =>
      queryClient.invalidateQueries({ queryKey: RESIDENT_LOCATION_GRAPHQL_QUERIES.all() }),
  });
}
