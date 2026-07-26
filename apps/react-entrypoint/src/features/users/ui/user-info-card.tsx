"use client";

import { Badge, Card, Group, Stack, Text, Title } from "@mantine/core";

interface UserInfoCardUser {
  id: string;
  name: string;
  email: string;
  emailVerified: boolean;
  image: string | null;
  createdAt: string;
  updatedAt: string;
}

interface UserInfoCardProps {
  user: UserInfoCardUser;
}

export function UserInfoCard({ user }: UserInfoCardProps) {
  return (
    <Card withBorder padding="lg" radius="md">
      <Stack gap="md">
        <Title order={3}>User Information</Title>

        <Group>
          <Text fw={500} w={140}>
            ID:
          </Text>
          <Text>{user.id}</Text>
        </Group>

        <Group>
          <Text fw={500} w={140}>
            Name:
          </Text>
          <Text>{user.name}</Text>
        </Group>

        <Group>
          <Text fw={500} w={140}>
            Email:
          </Text>
          <Text>{user.email}</Text>
        </Group>

        <Group>
          <Text fw={500} w={140}>
            Email Verified:
          </Text>
          <Badge color={user.emailVerified ? "green" : "gray"} variant="light">
            {user.emailVerified ? "Verified" : "Not verified"}
          </Badge>
        </Group>

        <Group>
          <Text fw={500} w={140}>
            Created:
          </Text>
          <Text>{user.createdAt}</Text>
        </Group>

        <Group>
          <Text fw={500} w={140}>
            Updated:
          </Text>
          <Text>{user.updatedAt}</Text>
        </Group>
      </Stack>
    </Card>
  );
}
