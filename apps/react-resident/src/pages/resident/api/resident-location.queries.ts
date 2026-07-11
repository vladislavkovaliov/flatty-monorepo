import { queryOptions, useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { getJson, postJson, putJson, deleteJson } from '../../../lib/utils';
import type {
  DtoCreateResidentLocationRequest,
  DtoDeleteResidentLocationResponse,
  DtoResidentLocationResponse,
  ResidentLocationListData,
} from '../../../lib/types/api';

export const RESIDENT_LOCATION_QUERIES = {
  all: () => ['resident-location'] as const,
  current: () =>
    queryOptions({
      queryKey: [...RESIDENT_LOCATION_QUERIES.all(), 'current'],
      queryFn: () => getJson<ResidentLocationListData>('/api/resident-location'),
    }),
};

export function useResidentLocation() {
  return useQuery(RESIDENT_LOCATION_QUERIES.current());
}

export interface ResidentAddressForm {
  country: string;
  city: string;
  apartment: string;
  house: string;
  street: string;
  postalCode: string;
}

export function toApiBody(data: ResidentAddressForm): DtoCreateResidentLocationRequest {
  return {
    country: data.country,
    city: data.city,
    apartment: data.apartment,
    house: data.house,
    street: data.street,
    postal_code: data.postalCode,
  };
}

export function useCreateResidentLocation() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (data: ResidentAddressForm) =>
      postJson<DtoResidentLocationResponse>('/api/resident-location', toApiBody(data)),
    onSettled: () => queryClient.invalidateQueries({ queryKey: RESIDENT_LOCATION_QUERIES.all() }),
  });
}

export function useUpdateResidentLocation() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: ({ id, data }: { id: number; data: ResidentAddressForm }) =>
      putJson<DtoResidentLocationResponse>(`/api/resident-location/${id}`, toApiBody(data)),
    onSettled: () => queryClient.invalidateQueries({ queryKey: RESIDENT_LOCATION_QUERIES.all() }),
  });
}

export function useDeleteResidentLocation() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (id: number) =>
      deleteJson<DtoDeleteResidentLocationResponse>(`/api/resident-location/${id}`),
    onSettled: () => queryClient.invalidateQueries({ queryKey: RESIDENT_LOCATION_QUERIES.all() }),
  });
}
