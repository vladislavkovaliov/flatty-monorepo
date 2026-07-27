import { Card, Group, Stack, Text, Title } from "@mantine/core";
import type { UserDetailLocation } from "../types";

interface ResidentLocationCardProps {
  location: UserDetailLocation | null;
}

export function ResidentLocationCard({ location }: ResidentLocationCardProps) {
  return (
    <Card withBorder padding="lg" radius="md">
      <Stack gap="md">
        <Title order={3}>Resident Location{location ? "s" : ""}</Title>

        {!location ? (
          <Text c="dimmed" size="sm">
            No resident locations found.
          </Text>
        ) : (
          <Card key={location.id.toString()} padding="sm" radius="sm">
            <Stack gap="xs">
              <Group>
                <Text fw={500} w={100}>
                  Address:
                </Text>
                <Text>
                  {location.street} {location.house}
                  {location.apartment ? `, apt. ${location.apartment}` : ""}
                </Text>
              </Group>

              <Group>
                <Text fw={500} w={100}>
                  City:
                </Text>
                <Text>{location.city}</Text>
              </Group>

              <Group>
                <Text fw={500} w={100}>
                  Country:
                </Text>
                <Text>{location.country}</Text>
              </Group>

              <Group>
                <Text fw={500} w={100}>
                  Postal Code:
                </Text>
                <Text>{location.postalCode}</Text>
              </Group>

              <Group>
                <Text fw={500} w={100}>
                  Created:
                </Text>
                <Text>{location.createdAt}</Text>
              </Group>

              <Group>
                <Text fw={500} w={100}>
                  Updated:
                </Text>
                <Text>{location.updatedAt}</Text>
              </Group>
            </Stack>
          </Card>
        )}
      </Stack>
    </Card>
  );
}
