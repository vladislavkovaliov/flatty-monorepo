"use client";

import {
  useDeleteExpense,
  useExpensesGraphql,
  useResidentLocationGraphql,
} from "@flatty-budget/sdk";
import type { ListExpenseResponse } from "@flatty-budget/sdk";
import { Box, Button, Container, Pagination, Table } from "@mantine/core";
import { useQueryClient } from "@tanstack/react-query";
import { usePathname, useRouter, useSearchParams } from "next/navigation";
import { useEffect } from "react";

const LIMIT = 10;

const MONTH: Record<number, string> = {
  1: "Jan",
  2: "Feb",
  3: "Mar",
  4: "Apr",
  5: "May",
  6: "Jun",
  7: "Jul",
  8: "Aug",
  9: "Sep",
  10: "Oct",
  11: "Nov",
  12: "Dec",
} as const;

type ExpenseRow = ListExpenseResponse["data"][number];

export function ExpensesTable() {
  const searchParams = useSearchParams() ?? new URLSearchParams();
  const router = useRouter();
  const pathname = usePathname();
  const page = Number(searchParams.get("page") || "1");
  const offset = (page - 1) * LIMIT;

  const { data: residentLocationsData } = useResidentLocationGraphql();
  const residentLocationId =
    residentLocationsData?.residentLocationList?.data?.[0]?.id;

  const { data } = useExpensesGraphql(
    residentLocationId as number,
    LIMIT,
    offset,
  );
  const queryClient = useQueryClient();
  const deleteMutation = useDeleteExpense();

  const handleDelete = (id: number) => {
    deleteMutation.mutate(id, {
      onSettled: () => {
        queryClient.invalidateQueries({
          queryKey: ["expense-stats", "graphql"],
        });
      },
    });
  };

  const total = data?.expenseList?.total ?? 0;
  const totalPages = Math.ceil(total / LIMIT);

  useEffect(() => {
    if (page > 1 && totalPages > 0 && page > totalPages) {
      const params = new URLSearchParams(searchParams.toString());
      params.set("page", "1");
      router.push(`${pathname}?${params.toString()}`);
    }
  });

  const handlePageChange = (newPage: number) => {
    const params = new URLSearchParams(searchParams.toString());
    params.set("page", String(newPage));
    router.push(`${pathname}?${params.toString()}`);
  };

  const rows = ((data?.expenseList?.data ?? []) as ExpenseRow[]).map((element) => (
    <Table.Tr key={element.id}>
      <Table.Td>{element.id}</Table.Td>
      <Table.Td>{element.amount}</Table.Td>
      <Table.Td>{element.description}</Table.Td>
      <Table.Td>{element.category?.description}</Table.Td>
      <Table.Td>{MONTH[element.month]}</Table.Td>
      <Table.Td>{element.year}</Table.Td>
      <Table.Td>
        <Button
          size="xs"
          variant="light"
          color="red"
          loading={deleteMutation.isPending}
          onClick={() => handleDelete(element.id)}
        >
          Delete
        </Button>
      </Table.Td>
    </Table.Tr>
  ));

  return (
    <>
      <Box py="md">
        <Container fluid>
          <Table>
            <Table.Thead>
              <Table.Tr>
                <Table.Th>ID</Table.Th>
                <Table.Th>Amount</Table.Th>
                <Table.Th>Description</Table.Th>
                <Table.Th>Category</Table.Th>
                <Table.Th>Month</Table.Th>
                <Table.Th>Year</Table.Th>
                <Table.Th>Actions</Table.Th>
              </Table.Tr>
            </Table.Thead>
            <Table.Tbody>{rows}</Table.Tbody>
          </Table>
        </Container>
      </Box>

      {totalPages > 1 ? (
        <Box py="xl">
          <Container fluid>
            <Pagination
              total={totalPages}
              value={page}
              onChange={handlePageChange}
            />
          </Container>
        </Box>
      ) : null}
    </>
  );
}
