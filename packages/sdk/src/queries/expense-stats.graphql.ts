import { queryOptions, useQuery } from '@tanstack/react-query';
import { graphqlRequest } from '../lib/graphql';
import type {
  ExpenseMonthlyAverage,
  ExpenseMonthlyTotal,
} from '../types/graphql';

const EXPENSE_MONTHLY_TOTALS = `
  query {
    expenseMonthlyTotals {
      data {
        month
        year
        totalSpent
        updatedAt
      }
    }
  }
`;

const EXPENSE_MONTHLY_AVERAGES = `
  query {
    expenseMonthlyAverages {
      data {
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
  totals: () =>
    queryOptions({
      queryKey: [...EXPENSE_STATS_GRAPHQL_QUERIES.all(), 'totals'],
      queryFn: () => graphqlRequest<GqlTotalsData>(EXPENSE_MONTHLY_TOTALS, {}),
    }),
  averages: () =>
    queryOptions({
      queryKey: [...EXPENSE_STATS_GRAPHQL_QUERIES.all(), 'averages'],
      queryFn: () => graphqlRequest<GqlAveragesData>(EXPENSE_MONTHLY_AVERAGES, {}),
    }),
};

export function useExpenseMonthlyTotalsGraphql() {
  return useQuery(EXPENSE_STATS_GRAPHQL_QUERIES.totals());
}

export function useExpenseMonthlyAveragesGraphql() {
  return useQuery(EXPENSE_STATS_GRAPHQL_QUERIES.averages());
}
