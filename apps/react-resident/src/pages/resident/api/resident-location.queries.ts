import { queryOptions, useQuery } from '@tanstack/react-query';
import { getJson } from '../../../lib/utils';
import type { IResidentLocationResponse } from '../model/types';
// import { apiClient } from '../../../shared/api/client';
// import type { IResidentLocationResponse } from '../model/types';

export const RESIDENT_LOCATION_QUERIES = {
  all: () => ['resident-location'] as const,
  current: () =>
    queryOptions({
      queryKey: [...RESIDENT_LOCATION_QUERIES.all(), 'current'],
      queryFn: () => getJson<IResidentLocationResponse>('/api/resident-location')
      // queryFn: async () => {
      //   const response = await fetch('/api/resident-location');
      //   const json = await response.json()
      //   console.log(json)
      //   return []
      // },
    }),
};

export function useResidentLocation() {
  return useQuery(RESIDENT_LOCATION_QUERIES.current());
}

