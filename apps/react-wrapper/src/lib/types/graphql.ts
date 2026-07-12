export type Maybe<T> = T | null;
export type InputMaybe<T> = Maybe<T>;
/** All built-in and custom scalars, mapped to their actual values */
export type Scalars = {
  ID: { input: string; output: string; }
  String: { input: string; output: string; }
  Boolean: { input: boolean; output: boolean; }
  Int: { input: number; output: number; }
  Float: { input: number; output: number; }
  /** A date-time string at UTC, such as 2019-12-03T09:54:33Z, compliant with the date-time format. */
  DateTime: { input: unknown; output: unknown; }
};

export type Category = {
  __typename?: 'Category';
  createdAt: Scalars['DateTime']['output'];
  description: Scalars['String']['output'];
  id: Scalars['Int']['output'];
  name: Scalars['String']['output'];
  updatedAt: Scalars['DateTime']['output'];
};

export type CategoryCountResponse = {
  __typename?: 'CategoryCountResponse';
  total: Scalars['Int']['output'];
};

export type CategoryInput = {
  description: Scalars['String']['input'];
  name: Scalars['String']['input'];
};

export type DeleteCategoryResponse = {
  __typename?: 'DeleteCategoryResponse';
  data: Scalars['Int']['output'];
};

export type DeleteResidentLocationResponse = {
  __typename?: 'DeleteResidentLocationResponse';
  data: Scalars['Int']['output'];
};

export type ListCategoryResponse = {
  __typename?: 'ListCategoryResponse';
  data: Array<Category>;
  total: Scalars['Int']['output'];
};

export type ListResidentLocationResponse = {
  __typename?: 'ListResidentLocationResponse';
  data: Array<ResidentLocation>;
  total: Scalars['Int']['output'];
};

export type Mutation = {
  __typename?: 'Mutation';
  createCategory: Category;
  createResidentLocation: ResidentLocation;
  deleteCategory: DeleteCategoryResponse;
  deleteResidentLocation: DeleteResidentLocationResponse;
  updateCategory: Category;
  updateResidentLocation: ResidentLocation;
};


export type MutationCreateCategoryArgs = {
  categoryData: CategoryInput;
};


export type MutationCreateResidentLocationArgs = {
  residentLocatoinData: ResidentLocationInput;
};


export type MutationDeleteCategoryArgs = {
  id: Scalars['Int']['input'];
};


export type MutationDeleteResidentLocationArgs = {
  id: Scalars['Int']['input'];
};


export type MutationUpdateCategoryArgs = {
  categoryData: CategoryInput;
  id: Scalars['Int']['input'];
};


export type MutationUpdateResidentLocationArgs = {
  id: Scalars['Int']['input'];
  residentLocatoinData: ResidentLocationInput;
};

export type Query = {
  __typename?: 'Query';
  categoryCount: CategoryCountResponse;
  categoryList: ListCategoryResponse;
  residentLocationCount: ResidentLocationCountResponse;
  residentLocationList: ListResidentLocationResponse;
};


export type QueryCategoryListArgs = {
  limit?: Scalars['Int']['input'];
  offset?: Scalars['Int']['input'];
};


export type QueryResidentLocationListArgs = {
  limit?: Scalars['Int']['input'];
  offset?: Scalars['Int']['input'];
};

export type ResidentLocation = {
  __typename?: 'ResidentLocation';
  apartment: Scalars['String']['output'];
  city: Scalars['String']['output'];
  country: Scalars['String']['output'];
  createdAt: Scalars['DateTime']['output'];
  house: Scalars['String']['output'];
  id: Scalars['Int']['output'];
  postalCode: Scalars['String']['output'];
  street: Scalars['String']['output'];
  updatedAt: Scalars['DateTime']['output'];
};

export type ResidentLocationCountResponse = {
  __typename?: 'ResidentLocationCountResponse';
  total: Scalars['Int']['output'];
};

export type ResidentLocationInput = {
  apartment: Scalars['String']['input'];
  city: Scalars['String']['input'];
  country: Scalars['String']['input'];
  house: Scalars['String']['input'];
  postalCode: Scalars['String']['input'];
  street: Scalars['String']['input'];
};
