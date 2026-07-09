import { queryOptions, useQuery } from '@tanstack/react-query';
import type { ListResidentLocationResponse } from '../../../lib/types/graphql';

const LIST_RESIDENT_LOCATION = `
  query ListResidentLocation($limit: Int, $offset: Int) {
    list(limit: $limit, offset: $offset) {
      data {
        id
        country
        city
        postalCode
        street
        house
        apartment
        createdAt
        updatedAt
      }
      total
    }
  }
`;

export const RESIDENT_LOCATION_GRAPHQL_QUERIES = {
  all: () => ['resident-location', 'graphql'] as const,
  list: (limit = 10, offset = 0) =>
    queryOptions({
      queryKey: [...RESIDENT_LOCATION_GRAPHQL_QUERIES.all(), 'list', { limit, offset }],
      queryFn: async () => {
        const response = await fetch('/graphql', {
          method: 'POST',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify({
            query: LIST_RESIDENT_LOCATION,
            variables: { limit, offset },
          }),
        });

        const json = await response.json();

        if (json.errors) {
          throw new Error(json.errors[0]?.message ?? 'GraphQL error');
        }

        return json.data.list as ListResidentLocationResponse;
      },
    }),
};

export function useResidentLocationGraphql(limit = 10, offset = 0) {
  return useQuery(RESIDENT_LOCATION_GRAPHQL_QUERIES.list(limit, offset));
}
