import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';
import { getJson, postJson, putJson, deleteJson } from './http';

describe('HTTP helpers', () => {
  const mockFetch = vi.fn();

  beforeEach(() => {
    vi.stubGlobal('fetch', mockFetch);
  });

  afterEach(() => {
    vi.restoreAllMocks();
  });

  describe('getJson', () => {
    it('makes a GET request with credentials', async () => {
      mockFetch.mockResolvedValueOnce({
        ok: true,
        json: () => Promise.resolve({ data: 'ok' }),
      });

      const result = await getJson<{ data: string }>('/api/test');

      expect(mockFetch).toHaveBeenCalledWith('/api/test', {
        method: 'GET',
        credentials: 'include',
      });
      expect(result).toEqual({ data: 'ok' });
    });

    it('throws on non-ok response with error message', async () => {
      mockFetch.mockResolvedValueOnce({
        ok: false,
        status: 400,
        json: () => Promise.resolve({ error: 'Bad request' }),
      });

      await expect(getJson('/api/test')).rejects.toThrow('Bad request');
    });

    it('throws generic message when error body is missing', async () => {
      mockFetch.mockResolvedValueOnce({
        ok: false,
        status: 500,
        json: () => Promise.resolve(null),
      });

      await expect(getJson('/api/test')).rejects.toThrow('Request failed (500)');
    });
  });

  describe('postJson', () => {
    it('sends POST with JSON body', async () => {
      mockFetch.mockResolvedValueOnce({
        ok: true,
        json: () => Promise.resolve({ id: 1 }),
      });

      const result = await postJson<{ id: number }>('/api/create', { name: 'test' });

      expect(mockFetch).toHaveBeenCalledWith('/api/create', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ name: 'test' }),
        credentials: 'include',
      });
      expect(result).toEqual({ id: 1 });
    });
  });

  describe('putJson', () => {
    it('sends PUT with JSON body', async () => {
      mockFetch.mockResolvedValueOnce({
        ok: true,
        json: () => Promise.resolve({ updated: true }),
      });

      const result = await putJson<{ updated: boolean }>('/api/update/1', { name: 'new' });

      expect(mockFetch).toHaveBeenCalledWith('/api/update/1', {
        method: 'PUT',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ name: 'new' }),
        credentials: 'include',
      });
      expect(result).toEqual({ updated: true });
    });
  });

  describe('deleteJson', () => {
    it('sends DELETE request', async () => {
      mockFetch.mockResolvedValueOnce({
        ok: true,
        json: () => Promise.resolve({ deleted: true }),
      });

      const result = await deleteJson<{ deleted: boolean }>('/api/delete/1');

      expect(mockFetch).toHaveBeenCalledWith('/api/delete/1', {
        method: 'DELETE',
        credentials: 'include',
      });
      expect(result).toEqual({ deleted: true });
    });
  });
});
