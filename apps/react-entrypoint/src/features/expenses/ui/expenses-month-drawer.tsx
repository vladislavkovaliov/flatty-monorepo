"use client";

import { useExpensesByYearMonth } from "@flatty-budget/sdk";
import {
  Alert,
  Badge,
  Button,
  Drawer,
  Group,
  Loader,
  SimpleGrid,
  Stack,
  Table,
  Text,
} from "@mantine/core";
import { useEffect, useState } from "react";

import type { MonthSummary } from "../lib/monthly-stats";

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

interface ExpensesMonthDrawerProps {
  opened: boolean;
  onClose: () => void;
  year: number;
  months: MonthSummary[];
  residentLocationId: number;
}

export function ExpensesMonthDrawer({
  opened,
  onClose,
  year,
  months,
  residentLocationId,
}: ExpensesMonthDrawerProps) {
  const [selectedMonth, setSelectedMonth] = useState<number | null>(null);

  useEffect(() => {
    setSelectedMonth(null);
  }, [year]);

  const { data, isPending, isError, error, refetch } = useExpensesByYearMonth(
    residentLocationId,
    year,
    selectedMonth ?? undefined,
  );

  const expenses = data?.data ?? [];

  return (
    <Drawer
      opened={opened}
      onClose={onClose}
      size="lg"
      padding="md"
      title={`Expenses — ${year}`}
    >
      <Stack gap="md">
        <SimpleGrid cols={4}>
          {months.map(({ month, expenseCount }) => {
            const isSelected = month === selectedMonth;
            return (
              <Button
                key={month}
                variant={isSelected ? "filled" : "light"}
                onClick={() => setSelectedMonth(month)}
              >
                <Group gap="xs" justify="center">
                  <Text>{MONTH[month]}</Text>
                  <Badge size="xs" variant="outline">
                    {expenseCount}
                  </Badge>
                </Group>
              </Button>
            );
          })}
        </SimpleGrid>

        {selectedMonth == null ? (
          <Text c="dimmed">Select a month to view its expenses.</Text>
        ) : isPending ? (
          <Group justify="center" py="xl">
            <Loader />
          </Group>
        ) : isError ? (
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
        ) : expenses.length === 0 ? (
          <Text c="dimmed">No expenses for this month.</Text>
        ) : (
          <Table>
            <Table.Thead>
              <Table.Tr>
                <Table.Th>Description</Table.Th>
                <Table.Th>Amount</Table.Th>
                <Table.Th>Month</Table.Th>
                <Table.Th>Year</Table.Th>
              </Table.Tr>
            </Table.Thead>
            <Table.Tbody>
              {expenses.map((expense) => (
                <Table.Tr key={expense.id}>
                  <Table.Td>{expense.description}</Table.Td>
                  <Table.Td>{expense.amount.toFixed(2)}</Table.Td>
                  <Table.Td>{MONTH[expense.month]}</Table.Td>
                  <Table.Td>{expense.year}</Table.Td>
                </Table.Tr>
              ))}
            </Table.Tbody>
          </Table>
        )}
      </Stack>
    </Drawer>
  );
}