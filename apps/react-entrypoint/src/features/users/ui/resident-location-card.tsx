import { Card, Group, Stack, Text, Title } from "@mantine/core";

interface ResidentLocationCardLocation {
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

interface ResidentLocationCardProps {
  locations: ResidentLocationCardLocation[];
}

export function ResidentLocationCard({ locations }: ResidentLocationCardProps) {
  return (
    <Card withBorder padding="lg" radius="md">
      <Stack gap="md">
        <Title order={3}>
          Resident Location{locations.length !== 1 ? "s" : ""}
        </Title>

        {locations.length === 0 ? (
          <Text c="dimmed" size="sm">
            No resident locations found.
          </Text>
        ) : (
          locations.map((location) => (
            <Card key={location.id} padding="sm" radius="sm">
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
          ))
        )}
      </Stack>
    </Card>
  );
}
