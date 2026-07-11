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

export interface DtoCountResponse {
  /** @example 138 */
  total: number;
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

export interface DtoDeleteResidentLocationResponse {
  /** @example 1 */
  data: number;
}

export interface DtoHealthResponse {
  /** @example "ok" */
  status: string;
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

export type HealthListData = DtoHealthResponse;

export interface ResidentLocationListParams {
  /** Number of products to return (default 10) */
  limit?: number;
  /** Number of products to skip (default 0) */
  offset?: number;
}

export type ResidentLocationListData = DtoListResidentLocationResponse;

export type ResidentLocationCreateData = DtoResidentLocationResponse;

export type CountListData = DtoCountResponse;

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
