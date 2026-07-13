import { queryOptions, useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { getJson, postJson, putJson, deleteJson } from '../lib/http';
import type {
  DtoDeleteResidentLocationResponse,
  CategoriesListData,
  DtoCreateCategoryRequest,
  DtoCategoryResponse,
  DtoDeleteCategoryResponse
} from '../types/api';

export const CATEGORIES_QUERIES = {
  all: () => ['categories'] as const,
  list: (limit: number, offset: number) =>
    queryOptions({
      queryKey: [...CATEGORIES_QUERIES.all(), 'list', { limit, offset }],
      queryFn: () => getJson<CategoriesListData>(`/api/categories?limit=${limit}&offset=${offset}`),
    }),
};

export function useCategories(limit = 10, offset = 0) {
  return useQuery(CATEGORIES_QUERIES.list(limit, offset));
}

export interface CategoriesForm {
  name: string;
  description: string;
}

export function toApiBody(data: CategoriesForm): DtoCreateCategoryRequest {
  return {
    name: data.name,
    description: data.description,
  };
}

export function useCreateCategories() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (data: CategoriesForm) =>
      postJson<DtoCategoryResponse>('/api/categories', toApiBody(data)),
    onSettled: () => queryClient.invalidateQueries({ queryKey: CATEGORIES_QUERIES.all() }),
  });
}

export function useUpdatecategories() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: ({ id, data }: { id: number; data: CategoriesForm }) =>
      putJson<DtoCategoryResponse>(`/api/categories/${id}`, toApiBody(data)),
    onSettled: () => queryClient.invalidateQueries({ queryKey: CATEGORIES_QUERIES.all() }),
  });
}

export function useDeleteCategory() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (id: number) =>
      deleteJson<DtoDeleteCategoryResponse>(`/api/categories/${id}`),
    onSettled: () => queryClient.invalidateQueries({ queryKey: CATEGORIES_QUERIES.all() }),
  });
}
