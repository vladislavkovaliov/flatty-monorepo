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

export type ListResidentLocationResponse = {
  __typename?: 'ListResidentLocationResponse';
  data: Array<ResidentLocation>;
  total: Scalars['Int']['output'];
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
