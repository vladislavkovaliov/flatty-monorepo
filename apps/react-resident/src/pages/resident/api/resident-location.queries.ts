import { queryOptions, useQuery } from '@tanstack/react-query';
import { getJson } from '../../../lib/utils';
import type { ResidentLocationListData } from '../../../lib/types/api';

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
