import { queryOptions, useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { getJson, postJson, deleteJson } from '../lib/http';
import type {
  ExpensesListData,
  ExpensesDeleteData,
  ExpensesCreateData,
  DtoCreateExpenseRequest,
  GetYearsAndMonthsListData,
  GetByYearMonthListData,
} from '../types/api';

export const EXPENSES_QUERIES = {
  all: () => ['expenses'] as const,
  list: (limit: number, offset: number) =>
    queryOptions({
      queryKey: [...EXPENSES_QUERIES.all(), 'list', { limit, offset }],
      queryFn: () => getJson<ExpensesListData>(`/api/expenses?limit=${limit}&offset=${offset}`),
    }),
  yearsAndMonths: (residentLocationId: number) =>
    queryOptions({
      queryKey: [...EXPENSES_QUERIES.all(), 'years-and-months', { residentLocationId }],
      queryFn: () =>
        getJson<GetYearsAndMonthsListData>(
          `/api/expenses/get-years-and-months?residentLocationId=${residentLocationId}`,
        ),
      enabled: residentLocationId != null,
    }),
  byYearMonth: (residentLocationId: number, year: number, month: number) =>
    queryOptions({
      queryKey: [
        ...EXPENSES_QUERIES.all(),
        'by-year-month',
        { residentLocationId, year, month },
      ],
      queryFn: () =>
        getJson<GetByYearMonthListData>(
          `/api/expenses/get-by-year-month?residentLocationId=${residentLocationId}&year=${year}&month=${month}`,
        ),
      enabled: residentLocationId != null && year != null && month != null,
    }),
};

export function useExpenses(limit = 10, offset = 0) {
  return useQuery(EXPENSES_QUERIES.list(limit, offset));
}

export function useExpensesYearsAndMonths(residentLocationId: number | undefined) {
  return useQuery({
    ...EXPENSES_QUERIES.yearsAndMonths(residentLocationId as number),
    enabled: residentLocationId != null,
  });
}

export function useExpensesByYearMonth(
  residentLocationId: number | undefined,
  year: number | undefined,
  month: number | undefined,
) {
  return useQuery(
    EXPENSES_QUERIES.byYearMonth(
      residentLocationId as number,
      year as number,
      month as number,
    ),
  );
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