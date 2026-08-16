import { describe, it, expect, vi } from 'vitest';
import { EXPENSES_QUERIES } from './expenses.queries';

describe('EXPENSES_QUERIES', () => {
  it('has a valid all() query key', () => {
    const key = EXPENSES_QUERIES.all();
    expect(key).toEqual(['expenses']);
  });

  it('list() has correct query key shape and queryFn', () => {
    const opts = EXPENSES_QUERIES.list(20, 0);
    expect(opts.queryKey).toEqual(['expenses', 'list', { limit: 20, offset: 0 }]);
    expect(typeof opts.queryFn).toBe('function');
  });

  it('yearsAndMonths() has correct query key shape and is enabled', () => {
    const opts = EXPENSES_QUERIES.yearsAndMonths(7);
    expect(opts.queryKey).toEqual([
      'expenses',
      'years-and-months',
      { residentLocationId: 7 },
    ]);
    expect(typeof opts.queryFn).toBe('function');
    expect(opts.enabled).toBe(true);
  });

  it('yearsAndMonths() disables when residentLocationId is null', () => {
    const opts = EXPENSES_QUERIES.yearsAndMonths(null as unknown as number);
    expect(opts.enabled).toBe(false);
  });

  it('yearsAndMonths() queryFn calls the correct URL', async () => {
    const fetchMock = vi.fn().mockResolvedValue({
      ok: true,
      json: async () => ({ data: [] }),
    });
    vi.stubGlobal('fetch', fetchMock);

    const opts = EXPENSES_QUERIES.yearsAndMonths(7);
    await opts.queryFn!({} as never);

    expect(fetchMock).toHaveBeenCalledWith(
      '/api/expenses/get-years-and-months?residentLocationId=7',
      expect.objectContaining({ method: 'GET', credentials: 'include' }),
    );
    vi.unstubAllGlobals();
  });

  it('byYearMonth() has correct query key shape and is enabled', () => {
    const opts = EXPENSES_QUERIES.byYearMonth(7, 2026, 8);
    expect(opts.queryKey).toEqual([
      'expenses',
      'by-year-month',
      { residentLocationId: 7, year: 2026, month: 8 },
    ]);
    expect(typeof opts.queryFn).toBe('function');
    expect(opts.enabled).toBe(true);
  });

  it('byYearMonth() disables when any param is missing', () => {
    const opts = EXPENSES_QUERIES.byYearMonth(
      7,
      undefined as unknown as number,
      8,
    );
    expect(opts.enabled).toBe(false);
  });

  it('byYearMonth() queryFn calls the correct URL', async () => {
    const fetchMock = vi.fn().mockResolvedValue({
      ok: true,
      json: async () => ({ data: [], total: 0 }),
    });
    vi.stubGlobal('fetch', fetchMock);

    const opts = EXPENSES_QUERIES.byYearMonth(7, 2026, 8);
    await opts.queryFn!({} as never);

    expect(fetchMock).toHaveBeenCalledWith(
      '/api/expenses/get-by-year-month?residentLocationId=7&year=2026&month=8',
      expect.objectContaining({ method: 'GET', credentials: 'include' }),
    );
    vi.unstubAllGlobals();
  });
});