import { queryOptions, useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import type {
  DeleteExpenseResponse,
  ListExpenseResponse,
  Expense,
  ExpenseInput,
} from '../types/graphql';
import { EXPENSES_QUERIES } from './expenses.queries';

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

const LIST_EXPENSE = `
query expenseList($limit: Int, $offset: Int) {
    expenseList(limit: $limit, offset: $offset) {
      data {
        id
        amount
        categoryId
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

const CREATE_EXPENSE = `
  mutation CreateExpense($input: ExpenseInput!) {
    createExpense(expenseData: $input) {
      id
      amount
      categoryId
      month
      year
    }
  }
`;

const UPDATE_EXPENSE = `
  mutation UpdateExpense($id: Int!, $input: ExpenseInput!) {
    updateExpense(id: $id, expenseData: $input) {
      id
      amount
      categoryId
      month
      year
    }
  }
`;

const DELETE_EXPENSE = `
  mutation DeleteExpense($id: Int!) {
    deleteExpense(id: $id) {
      data
    }
  }
`;

type GqlListData = { expenseList: ListExpenseResponse };
type GqlCreateData = { createExpense: Expense };
type GqlUpdateData = { updateExpense: Expense };
type GqlDeleteData = { deleteExpense: DeleteExpenseResponse };

export const EXPENSES_GRAPHQL_QUERIES = {
  all: () => ['expenses', 'graphql'] as const,
  list: (limit = 10, offset = 0) =>
    queryOptions({
      queryKey: [...EXPENSES_QUERIES.all(), 'list', { limit, offset }],
      queryFn: () => graphqlRequest<GqlListData>(LIST_EXPENSE, { limit, offset }),
    }),
};

export function useExpensesGraphql(limit = 10, offset = 0) {
  return useQuery(EXPENSES_GRAPHQL_QUERIES.list(limit, offset));
}

export function useCreateExpensesGraphql() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (input: ExpenseInput) =>
      graphqlRequest<GqlCreateData>(CREATE_EXPENSE, { input }),
    onSettled: () =>
      queryClient.invalidateQueries({ queryKey: EXPENSES_GRAPHQL_QUERIES.all() }),
  });
}

export function useUpdateExpensesGraphql() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: ({ id, data }: { id: number; data: ExpenseInput }) =>
      graphqlRequest<GqlUpdateData>(UPDATE_EXPENSE, { id, input: data }),
    onSettled: () =>
      queryClient.invalidateQueries({ queryKey: EXPENSES_QUERIES.all() }),
  });
}

export function useDeleteExpensesGraphql() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (id: number) =>
      graphqlRequest<GqlDeleteData>(DELETE_EXPENSE, { id }),
    onSettled: () =>
      queryClient.invalidateQueries({ queryKey: EXPENSES_QUERIES.all() }),
  });
}
