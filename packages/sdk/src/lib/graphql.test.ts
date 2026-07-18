import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';
import { graphqlRequest } from './graphql';

describe('graphqlRequest', () => {
  const mockFetch = vi.fn();

  beforeEach(() => {
    vi.stubGlobal('fetch', mockFetch);
  });

  afterEach(() => {
    vi.restoreAllMocks();
  });

  const query = `query { test { id } }`;
  const variables = { limit: 10 };

  it('sends a POST to /graphql with query and variables', async () => {
    mockFetch.mockResolvedValueOnce({
      ok: true,
      json: () => Promise.resolve({ data: { test: { id: 1 } } }),
    });

    const result = await graphqlRequest<{ test: { id: number } }>(query, variables);

    expect(mockFetch).toHaveBeenCalledWith('/graphql', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      credentials: 'include',
      body: JSON.stringify({ query, variables }),
    });
    expect(result).toEqual({ test: { id: 1 } });
  });

  it('throws on HTTP error with message from body', async () => {
    mockFetch.mockResolvedValueOnce({
      ok: false,
      status: 400,
      json: () => Promise.resolve({ message: 'Validation failed' }),
    });

    await expect(graphqlRequest(query, variables)).rejects.toThrow('Validation failed');
  });

  it('throws generic message on HTTP error without body', async () => {
    mockFetch.mockResolvedValueOnce({
      ok: false,
      status: 500,
      json: () => Promise.resolve(null),
    });

    await expect(graphqlRequest(query, variables)).rejects.toThrow('GraphQL request failed (500)');
  });

  it('throws first error from GraphQL errors array', async () => {
    mockFetch.mockResolvedValueOnce({
      ok: true,
      json: () => Promise.resolve({
        data: null,
        errors: [{ message: 'Resolver error' }, { message: 'Second error' }],
      }),
    });

    await expect(graphqlRequest(query, variables)).rejects.toThrow('Resolver error');
  });

  it('throws unknown error when errors array exists but is empty', async () => {
    mockFetch.mockResolvedValueOnce({
      ok: true,
      json: () => Promise.resolve({ data: null, errors: [] }),
    });

    await expect(graphqlRequest(query, variables)).rejects.toThrow('Unknown GraphQL error');
  });
});
