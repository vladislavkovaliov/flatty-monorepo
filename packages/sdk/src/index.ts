// HTTP helpers
export { getJson, postJson, putJson, deleteJson } from './lib/http';

// Generated API types
export type {
  DtoCategoryResponse,
  DtoCreateCategoryRequest,
  DtoCreateResidentLocationRequest,
  DtoDeleteCategoryResponse,
  DtoDeleteResidentLocationResponse,
  DtoHealthResponse,
  DtoListCategoryResponse,
  DtoListResidentLocationResponse,
  DtoResidentLocationResponse,
  DtoUpdateCategoryRequest,
  DtoUpdateResidentLocationRequest,
  CategoriesListParams,
  CategoriesListData,
  CategoriesCreateData,
  CountListData,
  CategoriesUpdateParams,
  CategoriesUpdateData,
  CategoriesDeleteParams,
  CategoriesDeleteData,
  HealthListData,
  ResidentLocationListParams,
  ResidentLocationListData,
  ResidentLocationCreateData,
  CountListResult,
  ResidentLocationUpdateParams,
  ResidentLocationUpdateData,
  ResidentLocationDeleteParams,
  ResidentLocationDeleteData,

  // New types for expenses
  DtoExpenseResponse,
  DtoListExpenseResponse,
  DtoCreateExpenseRequest,
  ExpensesListData,
  ExpensesCreateData,
  ExpensesDeleteData,
  ExpensesDeleteParams
} from './types/api';

// Generated GraphQL types
export type {
  Maybe,
  InputMaybe,
  Scalars,
  Category,
  CategoryCountResponse,
  CategoryInput,
  DeleteCategoryResponse,
  DeleteExpenseResponse,
  DeleteResidentLocationResponse,
  Expense,
  ExpenseInput,
  ListCategoryResponse,
  ListExpenseResponse,
  ListResidentLocationResponse,
  Mutation,
  Query,
  ResidentLocation,
  ResidentLocationCountResponse,
  ResidentLocationInput,
} from './types/graphql';

// REST queries
export {
  RESIDENT_LOCATION_QUERIES,
  useResidentLocation,
  useCreateResidentLocation,
  useUpdateResidentLocation,
  useDeleteResidentLocation,
} from './queries/resident-location.queries';
export type { ResidentAddressForm } from './queries/resident-location.queries';

export {
  CATEGORIES_QUERIES,
  useCategories,
  useCreateCategories,
  useUpdatecategories,
  useDeleteCategory,
} from './queries/categories.queries';
export type { CategoriesForm } from './queries/categories.queries';

export {
  EXPENSES_QUERIES,
  useExpenses,
  useCreateExpense,
  useDeleteExpense,
} from './queries/expenses.queries';

// GraphQL queries
export {
  RESIDENT_LOCATION_GRAPHQL_QUERIES,
  useResidentLocationGraphql,
  useCreateResidentLocationGraphql,
  useUpdateResidentLocationGraphql,
  useDeleteResidentLocationGraphql,
} from './queries/resident-location.graphql';

export {
  CATEGORIES_GRAPHQL_QUERIES,
  useCategoriesGraphql,
  useCreateCategoriesGraphql,
  useUpdateCategoryGraphql,
  useDeleteCategoryGraphql,
} from './queries/categories.graphql';

export {
  EXPENSES_GRAPHQL_QUERIES,
  useExpensesGraphql,
  useCreateExpensesGraphql,
  useUpdateExpensesGraphql,
  useDeleteExpensesGraphql,
} from './queries/expenses.graphql';
