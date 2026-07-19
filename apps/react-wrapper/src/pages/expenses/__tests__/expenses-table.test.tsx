//
// NOTE: This test requires Vitest + @testing-library/react to be configured.
// Currently the project has no test infrastructure installed.
// This file is provided as a scaffolding reference.
// To run: `npx vitest` (after vitest setup)
//

import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen, fireEvent } from '@testing-library/react';
import { MemoryRouter } from 'react-router-dom';
import { MantineProvider } from '@mantine/core';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { ExpensesTable } from '../ui/expenses-table';

// Mock the SDK hooks
vi.mock('@flatty-budget/sdk', () => ({
  useExpensesGraphql: vi.fn(),
  useDeleteExpense: vi.fn(),
}));

function createWrapper() {
  const queryClient = new QueryClient({
    defaultOptions: {
      queries: { retry: false },
      mutations: { retry: false },
    },
  });

  return function Wrapper({ children }: { children: React.ReactNode }) {
    return (
      <MantineProvider>
        <QueryClientProvider client={queryClient}>
          <MemoryRouter>
            {children}
          </MemoryRouter>
        </QueryClientProvider>
      </MantineProvider>
    );
  };
}

describe('ExpensesTable', () => {
  beforeEach(() => {
    vi.restoreAllMocks();
  });

  it('renders expense rows from GraphQL data', async () => {
    const { useExpensesGraphql } = await import('@flatty-budget/sdk');

    vi.mocked(useExpensesGraphql).mockReturnValue({
      data: {
        expenseList: {
          data: [
            { id: 1, amount: 100, category: { description: 'Food' }, month: 7, year: 2026 },
            { id: 2, amount: 50, category: { description: 'Transport' }, month: 7, year: 2026 },
          ],
          total: 2,
        },
      },
      isLoading: false,
      isError: false,
      error: null,
    } as any);

    const { useDeleteExpense } = await import('@flatty-budget/sdk');
    vi.mocked(useDeleteExpense).mockReturnValue({
      mutate: vi.fn(),
      isPending: false,
    } as any);


    render(<ExpensesTable />, { wrapper: createWrapper() });

    expect(screen.getByText('100')).toBeInTheDocument();
    expect(screen.getByText('Food')).toBeInTheDocument();
    expect(screen.getByText('50')).toBeInTheDocument();
    expect(screen.getByText('Transport')).toBeInTheDocument();
  });

  it('calls deleteMutation.mutate with the correct id when delete is clicked', async () => {
    const { useExpensesGraphql, useDeleteExpense } = await import('@flatty-budget/sdk');

    const mockMutate = vi.fn();
    vi.mocked(useDeleteExpense).mockReturnValue({
      mutate: mockMutate,
      isPending: false,
    } as any);

    vi.mocked(useExpensesGraphql).mockReturnValue({
      data: {
        expenseList: {
          data: [
            { id: 42, amount: 200, category: { description: 'Food' }, month: 7, year: 2026 },
          ],
          total: 1,
        },
      },
      isLoading: false,
      isError: false,
      error: null,
    } as any);


    render(<ExpensesTable />, { wrapper: createWrapper() });

    const deleteButton = screen.getByText('Delete');
    fireEvent.click(deleteButton);

    // The mutate should be called with the expense id (42) and an onSettled callback
    expect(mockMutate).toHaveBeenCalledWith(42, {
      onSettled: expect.any(Function),
    });
  });

  it('shows loading state on delete button when mutation is pending', async () => {
    const { useExpensesGraphql, useDeleteExpense } = await import('@flatty-budget/sdk');

    vi.mocked(useDeleteExpense).mockReturnValue({
      mutate: vi.fn(),
      isPending: true,
    } as any);

    vi.mocked(useExpensesGraphql).mockReturnValue({
      data: {
        expenseList: {
          data: [
            { id: 1, amount: 100, category: { description: 'Food' }, month: 7, year: 2026 },
          ],
          total: 1,
        },
      },
      isLoading: false,
      isError: false,
      error: null,
    } as any);


    render(<ExpensesTable />, { wrapper: createWrapper() });

    const deleteButton = screen.getByText('Delete');
    expect(deleteButton.closest('button')).toHaveAttribute('data-loading');
  });

  it('invalidates expense-stats cache on delete settled', async () => {
    const { useExpensesGraphql, useDeleteExpense } = await import('@flatty-budget/sdk');

    // Capture the onSettled callback
    let capturedOnSettled: (() => void) | undefined;
    const mockMutate = vi.fn((_id: number, options?: { onSettled?: () => void }) => {
      capturedOnSettled = options?.onSettled;
    });

    vi.mocked(useDeleteExpense).mockReturnValue({
      mutate: mockMutate,
      isPending: false,
    } as any);

    vi.mocked(useExpensesGraphql).mockReturnValue({
      data: {
        expenseList: {
          data: [
            { id: 1, amount: 100, category: { description: 'Food' }, month: 7, year: 2026 },
          ],
          total: 1,
        },
      },
      isLoading: false,
      isError: false,
      error: null,
    } as any);


    // Create a wrapper with a query client we can spy on
    const queryClient = new QueryClient({
      defaultOptions: {
        queries: { retry: false },
        mutations: { retry: false },
      },
    });
    const invalidateSpy = vi.spyOn(queryClient, 'invalidateQueries');

    function WrapperWithSpy({ children }: { children: React.ReactNode }) {
      return (
        <MantineProvider>
          <QueryClientProvider client={queryClient}>
            <MemoryRouter>
              {children}
            </MemoryRouter>
          </QueryClientProvider>
        </MantineProvider>
      );
    }

    render(<ExpensesTable />, { wrapper: WrapperWithSpy });

    const deleteButton = screen.getByText('Delete');
    fireEvent.click(deleteButton);

    // Extract the onSettled callback and invoke it
    const mutateCall = mockMutate.mock.calls[0];
    const onSettled = mutateCall[1]?.onSettled;
    expect(onSettled).toBeDefined();

    onSettled!();

    expect(invalidateSpy).toHaveBeenCalledWith({
      queryKey: ['expense-stats', 'graphql'],
    });
  });
});
