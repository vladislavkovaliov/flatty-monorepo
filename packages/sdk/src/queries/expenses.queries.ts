import { queryOptions, useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { getJson, postJson, deleteJson } from '../lib/http';
import type {
  ExpensesListData,
  ExpensesDeleteData,
  ExpensesCreateData,
  DtoCreateExpenseRequest,
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

export function useCreateExpense() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (input: DtoCreateExpenseRequest) =>
      postJson<ExpensesCreateData>('/api/expenses', input),
    onSettled: () => queryClient.invalidateQueries({ queryKey: EXPENSES_QUERIES.all() }),
  });
}

export function useDeleteExpense() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (id: number) =>
      deleteJson<ExpensesDeleteData>(`/api/expenses/${id}`),
    onSettled: () => queryClient.invalidateQueries({ queryKey: EXPENSES_QUERIES.all() }),
  });
}
