import { SimpleGrid, Card, Center, Text, Stack, Kbd } from '@mantine/core';
import { useNavigate } from 'react-router-dom';
import { APPS } from '../../../shared/config/apps';

export function HomePage() {
  const navigate = useNavigate();

  return (
    <Stack align="center" justify="center" h="100%" gap="xl">
      <Text size="xl" fw={700}>
        Applications
      </Text>

      <SimpleGrid cols={{ base: 1, sm: 2 }} spacing="lg">
        {APPS.map((app) => (
          <Card
            key={app.id}
            shadow="sm"
            padding="xl"
            radius="md"
            withBorder
            style={{ cursor: 'pointer', minWidth: 200 }}
            onClick={() => navigate(app.path)}
          >
            <Center>
              <Stack align="center" gap="sm">
                {app.icon}
                <Text fw={500} size="lg">
                  {app.label}
                </Text>
                <Text size="sm" c="dimmed">
                  {app.description}
                </Text>
              </Stack>
            </Center>
          </Card>
        ))}
      </SimpleGrid>

      <Text size="sm" c="dimmed">
        Press <Kbd>⌘</Kbd> + <Kbd>K</Kbd> to search
      </Text>
    </Stack>
  );
}
