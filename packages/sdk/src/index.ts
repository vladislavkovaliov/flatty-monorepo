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
  DeleteResidentLocationResponse,
  ListCategoryResponse,
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
