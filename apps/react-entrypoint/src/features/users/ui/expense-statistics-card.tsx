import { Card, Group, Stack, Text, Title } from "@mantine/core";
import type {
  UserDetailExpenseAverage,
  UserDetailExpenseTotal,
} from "../types";

interface ExpenseStatisticsCardProps {
  total: UserDetailExpenseTotal | null;
  average: UserDetailExpenseAverage | null;
  hasExpenseData: boolean;
}

export function ExpenseStatisticsCard({
  total,
  average,
  hasExpenseData,
}: ExpenseStatisticsCardProps) {
  return (
    <Card withBorder padding="lg" radius="md">
      <Stack gap="md">
        <Title order={3}>Expense Statistics</Title>

        {!hasExpenseData ? (
          <Text c="dimmed" size="sm">
            No expense data available.
          </Text>
        ) : (
          <>
            {total && (
              <SubCard title="Latest Monthly Total">
                <Group>
                  <Text fw={500} w={140}>
                    Period:
                  </Text>
                  <Text>
                    {total.month}/{total.year}
                  </Text>
                </Group>
                <Group>
                  <Text fw={500} w={140}>
                    Total Spent:
                  </Text>
                  <Text>{total.totalSpent}</Text>
                </Group>
              </SubCard>
            )}

            {average && (
              <SubCard title="Latest Monthly Average">
                <Group>
                  <Text fw={500} w={140}>
                    Period:
                  </Text>
                  <Text>
                    {average.month}/{average.year}
                  </Text>
                </Group>
                <Group>
                  <Text fw={500} w={140}>
                    Average Amount:
                  </Text>
                  <Text>{average.averageAmount}</Text>
                </Group>
                <Group>
                  <Text fw={500} w={140}>
                    Expense Count:
                  </Text>
                  <Text>{average.expenseCount}</Text>
                </Group>
              </SubCard>
            )}
          </>
        )}
      </Stack>
    </Card>
  );
}

function SubCard({
  title,
  children,
}: {
  title: string;
  children: React.ReactNode;
}) {
  return (
    <Card withBorder padding="sm" radius="sm">
      <Title order={5} mb="xs">
        {title}
      </Title>
      <Stack gap="xs">{children}</Stack>
    </Card>
  );
}
