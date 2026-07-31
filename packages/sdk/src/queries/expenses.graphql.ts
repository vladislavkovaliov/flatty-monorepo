import { queryOptions, useQuery } from '@tanstack/react-query';
import type { ListExpenseResponse } from '../types/graphql';
import { graphqlRequest } from '../lib/graphql';


const LIST_EXPENSE = `
query expenseList($residentLocationId: Int!, $limit: Int, $offset: Int) {
    expenseList(residentLocationId: $residentLocationId, limit: $limit, offset: $offset) {
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
  list: (residentLocationId: number, limit = 10, offset = 0) =>
    queryOptions({
      queryKey: [...EXPENSES_GRAPHQL_QUERIES.all(), 'list', { residentLocationId, limit, offset }],
      queryFn: () => graphqlRequest<GqlListData>(LIST_EXPENSE, { residentLocationId, limit, offset }),
    }),
};

export function useExpensesGraphql(residentLocationId: number, limit = 10, offset = 0) {
  console.log(residentLocationId)
  return useQuery(EXPENSES_GRAPHQL_QUERIES.list(residentLocationId, limit, offset));
}
