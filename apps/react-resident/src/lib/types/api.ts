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

export interface DtoCategoryResponse {
  /** @example "2026-07-09 08:34:05.796617" */
  created_at: string;
  /** @example "Коммунальные платежи" */
  description: string;
  /** @example 1 */
  id: number;
  /** @example "utilities" */
  name: string;
  /** @example "2026-07-09 08:34:05.796617" */
  updated_at: string;
}

export interface DtoCountResponse {
  /** @example 138 */
  total: number;
}

export interface DtoCreateCategoryRequest {
  /** @example "Коммунальные платежи" */
  description: string;
  /** @example "utilities" */
  name: string;
}

export interface DtoCreateResidentLocationRequest {
  /** @example "2" */
  apartment: string;
  /** @example "Warsaw" */
  city: string;
  /** @example "Poland" */
  country: string;
  /** @example "1" */
  house: string;
  /** @example "00-945" */
  postal_code: string;
  /** @example "Bobr" */
  street: string;
}

export interface DtoDeleteCategoryResponse {
  /** @example 1 */
  data: number;
}

export interface DtoDeleteResidentLocationResponse {
  /** @example 1 */
  data: number;
}

export interface DtoHealthResponse {
  /** @example "ok" */
  status: string;
}

export interface DtoListCategoryResponse {
  data: DtoCategoryResponse[];
  total: number;
}

export interface DtoListResidentLocationResponse {
  data: DtoResidentLocationResponse[];
  total: number;
}

export interface DtoResidentLocationResponse {
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

export interface DtoUpdateCategoryRequest {
  /** @example "Коммунальные платежи" */
  description: string;
  /** @example "utilities" */
  name: string;
}

export interface DtoUpdateResidentLocationRequest {
  /** @example "2" */
  apartment: string;
  /** @example "Warsaw" */
  city: string;
  /** @example "Poland" */
  country: string;
  /** @example "1" */
  house: string;
  /** @example "00-945" */
  postal_code: string;
  /** @example "Bobr" */
  street: string;
}

export interface CategoriesListParams {
  /** Number of products to return (default 10) */
  limit?: number;
  /** Number of products to skip (default 0) */
  offset?: number;
}

export type CategoriesListData = DtoListCategoryResponse;

export type CategoriesCreateData = DtoCategoryResponse;

export type CountListData = DtoCountResponse;

export interface CategoriesUpdateParams {
  /** Category ID */
  id: number;
}

export type CategoriesUpdateData = DtoCategoryResponse;

export interface CategoriesDeleteParams {
  /** Category ID */
  id: number;
}

export type CategoriesDeleteData = DtoDeleteCategoryResponse;

export type HealthListData = DtoHealthResponse;

export interface ResidentLocationListParams {
  /** Number of products to return (default 10) */
  limit?: number;
  /** Number of products to skip (default 0) */
  offset?: number;
}

export type ResidentLocationListData = DtoListResidentLocationResponse;

export type ResidentLocationCreateData = DtoResidentLocationResponse;

export type CountListResult = DtoCountResponse;

export interface ResidentLocationUpdateParams {
  /** Resident Location ID */
  id: number;
}

export type ResidentLocationUpdateData = DtoResidentLocationResponse;

export interface ResidentLocationDeleteParams {
  /** Resident Location ID */
  id: number;
}

export type ResidentLocationDeleteData = DtoDeleteResidentLocationResponse;
