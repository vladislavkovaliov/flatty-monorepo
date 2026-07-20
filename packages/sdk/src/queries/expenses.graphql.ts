import { queryOptions, useQuery } from '@tanstack/react-query';
import type { ListExpenseResponse } from '../types/graphql';
import { graphqlRequest } from '../lib/graphql';


const LIST_EXPENSE = `
query expenseList($limit: Int, $offset: Int) {
    expenseList(limit: $limit, offset: $offset) {
      data {
        id
        amount
        categoryId
        description
        category {
          id
          name
          description
        }
        month
        year
        residentLocationId
        createdAt
        updatedAt
      }
      total
    }
  }
`;

type GqlListData = { expenseList: ListExpenseResponse };

export const EXPENSES_GRAPHQL_QUERIES = {
  all: () => ['expenses', 'graphql'] as const,
  list: (limit = 10, offset = 0) =>
    queryOptions({
      queryKey: [...EXPENSES_GRAPHQL_QUERIES.all(), 'list', { limit, offset }],
      queryFn: () => graphqlRequest<GqlListData>(LIST_EXPENSE, { limit, offset }),
    }),
};

export function useExpensesGraphql(limit = 10, offset = 0) {
  return useQuery(EXPENSES_GRAPHQL_QUERIES.list(limit, offset));
}
