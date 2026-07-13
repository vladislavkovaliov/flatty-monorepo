import { queryOptions, useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { getJson, deleteJson } from '../lib/http';
import type {
  ExpensesListData,
  ExpensesDeleteData,
} from '../types/api';

export const EXPENSES_QUERIES = {
  all: () => ['expenses'] as const,
  list: (limit: number, offset: number) =>
    queryOptions({
      queryKey: [...EXPENSES_QUERIES.all(), 'list', { limit, offset }],
      queryFn: () => getJson<ExpensesListData>(`/api/expenses?limit=${limit}&offset=${offset}`),
    }),
};

export function useExpenses(limit = 10, offset = 0) {
  return useQuery(EXPENSES_QUERIES.list(limit, offset));
}

export function useDeleteExpense() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (id: number) =>
      deleteJson<ExpensesDeleteData>(`/api/expenses/${id}`),
    onSettled: () => queryClient.invalidateQueries({ queryKey: EXPENSES_QUERIES.all() }),
  });
}
