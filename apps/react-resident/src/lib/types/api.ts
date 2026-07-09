/* eslint-disable */
/* tslint:disable */
// @ts-nocheck
/*
 * ---------------------------------------------------------------
 * ## THIS FILE WAS GENERATED VIA SWAGGER-TYPESCRIPT-API        ##
 * ##                                                           ##
 * ## AUTHOR: acacode                                           ##
 * ## SOURCE: https://github.com/acacode/swagger-typescript-api ##
 * ---------------------------------------------------------------
 */

export interface FlattyBudgetGoApiHttpDtoCountResponse {
  /** @example 138 */
  total: number;
}

export interface FlattyBudgetGoApiHttpDtoHealthResponse {
  /** @example "ok" */
  status: string;
}

export interface FlattyBudgetGoApiHttpDtoListResidentLocationResponse {
  data: FlattyBudgetGoApiHttpDtoResidentLocationResponse[];
  total: number;
}

export interface FlattyBudgetGoApiHttpDtoResidentLocationResponse {
  /** @example "2" */
  apartment: string;
  /** @example "Warsaw" */
  city: string;
  /** @example "Poland" */
  country: string;
  /** @example "2026-07-09 08:34:05.796617" */
  created_at: string;
  /** @example "1" */
  house: string;
  /** @example 123 */
  id: number;
  /** @example "00-945" */
  postal_code: string;
  /** @example "Bobr" */
  street: string;
  /** @example "2026-07-09 08:34:05.796617" */
  updated_at: string;
}

export type HealthListData = FlattyBudgetGoApiHttpDtoHealthResponse;

export interface ResidentLocationListParams {
  /** Number of products to return (default 10) */
  limit?: number;
  /** Number of products to skip (default 0) */
  offset?: number;
}

export type ResidentLocationListData =
  FlattyBudgetGoApiHttpDtoListResidentLocationResponse;

export type CountListData = FlattyBudgetGoApiHttpDtoCountResponse;
