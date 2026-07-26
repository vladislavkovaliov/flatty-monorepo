import { Card, Group, Stack, Text, Title } from "@mantine/core";

interface ExpenseTotal {
  month: number;
  year: number;
  totalSpent: string;
}

interface ExpenseAverage {
  month: number;
  year: number;
  averageAmount: string;
  expenseCount: number;
}

interface ExpenseStatisticsCardProps {
  total: ExpenseTotal[];
  average: ExpenseAverage[];
}

export function ExpenseStatisticsCard({
  total,
  average,
}: ExpenseStatisticsCardProps) {
  const hasData = total.length > 0 || average.length > 0;

  return (
    <Card withBorder padding="lg" radius="md">
      <Stack gap="md">
        <Title order={3}>Expense Statistics</Title>

        {!hasData ? (
          <Text c="dimmed" size="sm">
            No expense data available.
          </Text>
        ) : (
          <>
            {total[0] && (
              <SubCard title="Latest Monthly Total">
                <Group>
                  <Text fw={500} w={140}>
                    Period:
                  </Text>
                  <Text>
                    {total[0].month}/{total[0].year}
                  </Text>
                </Group>
                <Group>
                  <Text fw={500} w={140}>
                    Total Spent:
                  </Text>
                  <Text>{total[0].totalSpent}</Text>
                </Group>
              </SubCard>
            )}

            {average[0] && (
              <SubCard title="Latest Monthly Average">
                <Group>
                  <Text fw={500} w={140}>
                    Period:
                  </Text>
                  <Text>
                    {average[0].month}/{average[0].year}
                  </Text>
                </Group>
                <Group>
                  <Text fw={500} w={140}>
                    Average Amount:
                  </Text>
                  <Text>{average[0].averageAmount}</Text>
                </Group>
                <Group>
                  <Text fw={500} w={140}>
                    Expense Count:
                  </Text>
                  <Text>{average[0].expenseCount}</Text>
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
