import { Card, Container, Stack, Text } from "@mantine/core";
import { ExpenseStatisticsCard } from "./expense-statistics-card";
import { ResidentLocationCard } from "./resident-location-card";
import { UserInfoCard } from "./user-info-card";

interface UserDetailUser {
  id: string;
  name: string;
  email: string;
  emailVerified: boolean;
  image: string | null;
  createdAt: string;
  updatedAt: string;
}

interface UserDetailLocation {
  id: number;
  country: string;
  city: string;
  postalCode: string;
  street: string;
  house: string;
  apartment: string | null;
  createdAt: string | null;
  updatedAt: string | null;
}

interface UserDetailExpenseTotal {
  month: number;
  year: number;
  totalSpent: string;
}

interface UserDetailExpenseAverage {
  month: number;
  year: number;
  averageAmount: string;
  expenseCount: number;
}

interface UserDetailProps {
  user: UserDetailUser | null;
  locations: UserDetailLocation[];
  expenseTotal: UserDetailExpenseTotal[];
  expenseAverage: UserDetailExpenseAverage[];
}

export function UserDetail({
  user,
  locations,
  expenseTotal,
  expenseAverage,
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
        <ResidentLocationCard locations={locations} />
        <ExpenseStatisticsCard total={expenseTotal} average={expenseAverage} />
      </Stack>
    </Container>
  );
}
