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

export type DeleteResidentLocationResponse = {
  __typename?: 'DeleteResidentLocationResponse';
  data: Scalars['Int']['output'];
};

export type ListResidentLocationResponse = {
  __typename?: 'ListResidentLocationResponse';
  data: Array<ResidentLocation>;
  total: Scalars['Int']['output'];
};

export type Mutation = {
  __typename?: 'Mutation';
  create: ResidentLocation;
  delete: DeleteResidentLocationResponse;
  update: ResidentLocation;
};


export type MutationCreateArgs = {
  residentLocatoinData: ResidentLocationInput;
};


export type MutationDeleteArgs = {
  id: Scalars['Int']['input'];
};


export type MutationUpdateArgs = {
  id: Scalars['Int']['input'];
  residentLocatoinData: ResidentLocationInput;
};

export type Query = {
  __typename?: 'Query';
  count: ResidentLocationCountResponse;
  list: ListResidentLocationResponse;
};


export type QueryListArgs = {
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
