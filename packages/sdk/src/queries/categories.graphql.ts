import { queryOptions, useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import type {
  DeleteCategoryResponse,
  ListCategoryResponse,
  Category,
  CategoryInput, 
} from '../types/graphql';
import { CATEGORIES_QUERIES } from './categories.queries';

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

const LIST_CATEGORY = `
query categoryList($limit: Int, $offset: Int) {
    list(limit: $limit, offset: $offset) {
      data {
        updatedAt
        name
        id
        description
        createdAt
      }
      total
    }
  }
`

const CREATE_CATEGORY = `
  mutation CreateCategory($input: CategoryInput!) {
    create(categoryData: $input) {
      name
      description
    }
  }
`;

const UPDATE_CATEGORY = `
  mutation UpdateCategory($id: Int!, $input: CategoryInput!) {
    update(id: $id, categoryData: $input) {
      id
      name
      description
    }
  }
`;


const DELETE_CATEGORY = `
  mutation DeleteCategory($id: Int!) {
    delete(id: $id) {
      data
    }
  }
`;

type GqlListData = { list: ListCategoryResponse };
type GqlCreateData = { create: Category };
type GqlUpdateData = { update: Category };
type GqlDeleteData = { delete: DeleteCategoryResponse };

export const CATEGORIES_GRAPHQL_QUERIES = {
  all: () => ['categories', 'graphql'] as const,
  list: (limit = 10, offset = 0) =>
    queryOptions({
      queryKey: [...CATEGORIES_QUERIES.all(), 'list', { limit, offset }],
      queryFn: () => graphqlRequest<GqlListData>(LIST_CATEGORY, { limit, offset }),
    }),
};

export function useCategoriesGraphql(limit = 10, offset = 0) {
  return useQuery(CATEGORIES_GRAPHQL_QUERIES.list(limit, offset));
}

export function useCreateCategoriesGraphql() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (input: CategoryInput) =>
      graphqlRequest<GqlCreateData>(CREATE_CATEGORY, { input }),
    onSettled: () =>
      queryClient.invalidateQueries({ queryKey: CATEGORIES_GRAPHQL_QUERIES.all() }),
  });
}

export function useUpdateCategoryGraphql() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: ({ id, data }: { id: number; data: CategoryInput }) =>
      graphqlRequest<GqlUpdateData>(UPDATE_CATEGORY, { id, input: data }),
    onSettled: () =>
      queryClient.invalidateQueries({ queryKey: CATEGORIES_QUERIES.all() }),
  });
}

export function useDeleteCategoryGraphql() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (id: number) =>
      graphqlRequest<GqlDeleteData>(DELETE_CATEGORY, { id }),
    onSettled: () =>
      queryClient.invalidateQueries({ queryKey: CATEGORIES_QUERIES.all() }),
  });
}
