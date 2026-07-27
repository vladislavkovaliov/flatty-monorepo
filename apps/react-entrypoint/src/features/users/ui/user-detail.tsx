import { Card, Container, Stack, Text } from "@mantine/core";
import type {
  UserDetailExpenseAverage,
  UserDetailExpenseTotal,
  UserDetailLocation,
  UserDetailUser,
} from "../types";
import { ExpenseStatisticsCard } from "./expense-statistics-card";
import { ResidentLocationCard } from "./resident-location-card";
import { UserInfoCard } from "./user-info-card";

interface UserDetailProps {
  user: UserDetailUser | null;
  locations: UserDetailLocation | null;
  expenseTotal: UserDetailExpenseTotal | null;
  expenseAverage: UserDetailExpenseAverage | null;
  hasExpenseData: boolean;
}

export function UserDetail({
  user,
  locations,
  expenseTotal,
  expenseAverage,
  hasExpenseData,
}: UserDetailProps) {
  if (!user) {
    return (
      <Container size="sm" py="xl">
        <Card withBorder padding="lg" radius="md">
          <Text c="dimmed">User not found.</Text>
        </Card>
      </Container>
    );
  }

  return (
    <Container size="sm" py="xl">
      <Stack gap="lg">
        <UserInfoCard user={user} />
        <ResidentLocationCard location={locations} />
        <ExpenseStatisticsCard
          hasExpenseData={hasExpenseData}
          total={expenseTotal}
          average={expenseAverage}
        />
      </Stack>
    </Container>
  );
}
