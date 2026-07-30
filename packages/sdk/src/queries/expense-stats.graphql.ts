import { queryOptions, useQuery } from '@tanstack/react-query';
import { graphqlRequest } from '../lib/graphql';
import type {
  ExpenseMonthlyAverage,
  ExpenseMonthlyTotal,
} from '../types/graphql';

const EXPENSE_MONTHLY_TOTALS = `
  query ExpenseMonthlyTotals($residentLocationId: Int!, $month: Int, $year: Int) {
    expenseMonthlyTotals(residentLocationId: $residentLocationId, month: $month, year: $year) {
      data {
        residentLocationId
        month
        year
        totalSpent
        updatedAt
      }
    }
  }
`;

const EXPENSE_MONTHLY_AVERAGES = `
  query ExpenseMonthlyAverages($residentLocationId: Int!, $month: Int, $year: Int) {
    expenseMonthlyAverages(residentLocationId: $residentLocationId, month: $month, year: $year) {
      data {
        residentLocationId
        month
        year
        averageAmount
        expenseCount
        updatedAt
      }
    }
  }
`;

type GqlTotalsData = { expenseMonthlyTotals: { data: ExpenseMonthlyTotal[] } };
type GqlAveragesData = { expenseMonthlyAverages: { data: ExpenseMonthlyAverage[] } };

export const EXPENSE_STATS_GRAPHQL_QUERIES = {
  all: () => ['expense-stats', 'graphql'] as const,
  totals: (residentLocationId: number, month?: number, year?: number) =>
    queryOptions({
      queryKey: [...EXPENSE_STATS_GRAPHQL_QUERIES.all(), 'totals', { residentLocationId, month, year }],
      queryFn: () =>
        graphqlRequest<GqlTotalsData>(EXPENSE_MONTHLY_TOTALS, { residentLocationId, month, year }),
      enabled: residentLocationId != null,
    }),
  averages: (residentLocationId: number, month?: number, year?: number) =>
    queryOptions({
      queryKey: [...EXPENSE_STATS_GRAPHQL_QUERIES.all(), 'averages', { residentLocationId, month, year }],
      queryFn: () =>
        graphqlRequest<GqlAveragesData>(EXPENSE_MONTHLY_AVERAGES, { residentLocationId, month, year }),
      enabled: residentLocationId != null,
    }),
};

export function useExpenseMonthlyTotalsGraphql(
  residentLocationId: number | undefined,
  month?: number,
  year?: number,
) {
  return useQuery({
    ...EXPENSE_STATS_GRAPHQL_QUERIES.totals(residentLocationId as number, month, year),
    enabled: residentLocationId != null,
  });
}

export function useExpenseMonthlyAveragesGraphql(
  residentLocationId: number | undefined,
  month?: number,
  year?: number,
) {
  return useQuery({
    ...EXPENSE_STATS_GRAPHQL_QUERIES.averages(residentLocationId as number, month, year),
    enabled: residentLocationId != null,
  });
}
