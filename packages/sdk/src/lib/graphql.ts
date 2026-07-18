/**
 * Generic GraphQL request helper.
 * Sends a POST request to /graphql with credentials included.
 */
export async function graphqlRequest<T>(
  query: string,
  variables: Record<string, unknown>,
): Promise<T> {
  const response = await fetch('/graphql', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    credentials: 'include',
    body: JSON.stringify({ query, variables }),
  });

  if (!response.ok) {
    const body = await response.json().catch(() => null);
    throw new Error(body?.message ?? `GraphQL request failed (${response.status})`);
  }

  const json = await response.json();

  if (json.errors) {
    const message = json.errors[0]?.message ?? 'Unknown GraphQL error';
    throw new Error(message);
  }

  return json.data as T;
}
