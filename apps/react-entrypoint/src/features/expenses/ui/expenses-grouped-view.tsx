"use client";

import type { GetByYearMonthListData } from "@flatty-budget/sdk";
import {
  useDeleteExpense,
  useExpensesByYearMonth,
  useExpensesYearsAndMonths,
  useResidentLocationGraphql,
} from "@flatty-budget/sdk";

import {
  Accordion,
  Alert,
  Box,
  Button,
  Container,
  Group,
  Loader,
  Stack,
  Table,
  Text,
} from "@mantine/core";
import { useMemo } from "react";

import { groupYearsAndMonths } from "../lib/monthly-stats";

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
};

interface MonthExpensesTableProps {
  residentLocationId: number | undefined;
  year: number;
  month: number;
}

type ExpenseRow = GetByYearMonthListData["data"][number];

function MonthExpensesTable({
  residentLocationId,
  year,
  month,
}: MonthExpensesTableProps) {
  const { data: expenses, isPending } = useExpensesByYearMonth(
    residentLocationId,
    year,
    month,
  );

  const deleteMutation = useDeleteExpense();

  if (isPending) {
    return (
      <Group justify="center" py="md">
        <Loader />
      </Group>
    );
  }

  const rows = ((expenses?.data ?? []) as ExpenseRow[]).map((element) => (
    <Table.Tr key={element.id}>
      <Table.Td>{element.id}</Table.Td>
      <Table.Td>{element.amount}</Table.Td>
      <Table.Td>{element.description}</Table.Td>
      <Table.Td>{element.category?.description}</Table.Td>
      <Table.Td>
        <Button
          size="xs"
          variant="light"
          color="red"
          loading={
            deleteMutation.isPending && deleteMutation.variables === element.id
          }
          onClick={() => deleteMutation.mutate(element.id)}
        >
          Delete
        </Button>
      </Table.Td>
    </Table.Tr>
  ));

  return (
    <Table>
      <Table.Thead>
        <Table.Tr>
          <Table.Th>ID</Table.Th>
          <Table.Th>Amount</Table.Th>
          <Table.Th>Description</Table.Th>
          <Table.Th>Category</Table.Th>
          <Table.Th>Actions</Table.Th>
        </Table.Tr>
      </Table.Thead>
      <Table.Tbody>{rows}</Table.Tbody>
    </Table>
  );
}

export function ExpensesGroupedView() {
  const { data: residentLocationsData } = useResidentLocationGraphql();
  const residentLocationId =
    residentLocationsData?.residentLocationList?.data?.[0]?.id;

  const { data, isPending, isError, error, refetch } =
    useExpensesYearsAndMonths(residentLocationId);

  const yearGroups = useMemo(
    () => groupYearsAndMonths(data?.data ?? []),
    [data?.data],
  );

  if (isPending) {
    return (
      <Box py="md">
        <Container fluid>
          <Group justify="center" py="xl">
            <Loader />
          </Group>
        </Container>
      </Box>
    );
  }

  if (isError) {
    return (
      <Box py="md">
        <Container fluid>
          <Alert color="red" title="Failed to load expenses">
            <Stack gap="sm">
              <Text>{error?.message ?? "Something went wrong."}</Text>
              <Button
                size="xs"
                variant="light"
                color="red"
                onClick={() => refetch()}
              >
                Retry
              </Button>
            </Stack>
          </Alert>
        </Container>
      </Box>
    );
  }

  if (yearGroups.length === 0) {
    return (
      <Box py="md">
        <Container fluid>
          <Text c="dimmed">No expenses yet.</Text>
        </Container>
      </Box>
    );
  }

  return (
    <Box py="md">
      <Container fluid>
        <Accordion multiple keepMounted>
          {yearGroups.map((group) => {
            return (
              <Accordion.Item key={group.year} value={group.year.toString()}>
                <Accordion.Control>{group.year}</Accordion.Control>
                <Accordion.Panel>
                  <Accordion multiple keepMounted>
                    {group.months.map((m) => {
                      return (
                        <Accordion.Item
                          key={`${group.year}-${m.month}`}
                          value={`${group.year}-${m.month}`}
                        >
                          <Accordion.Control>
                            {MONTH[m.month]}
                          </Accordion.Control>
                          <Accordion.Panel>
                            <Box py="md">
                              <Container fluid>
                                <MonthExpensesTable
                                  residentLocationId={residentLocationId}
                                  year={group.year}
                                  month={m.month}
                                />
                              </Container>
                            </Box>
                          </Accordion.Panel>
                        </Accordion.Item>
                      );
                    })}
                  </Accordion>
                </Accordion.Panel>
              </Accordion.Item>
            );
          })}
        </Accordion>
      </Container>
    </Box>
  );
}
