import { SimpleGrid, Card, Center, Text, Stack, Kbd } from '@mantine/core';
import { useNavigate } from 'react-router-dom';
import { APPS, type AppDefinition } from '../../../shared/config/apps';

export function HomePage() {
  const navigate = useNavigate();

  const handleNavigateCallback = (app: AppDefinition) => () => {

    const backendServices = ['openapi', 'graphql'];

     if (backendServices.includes(app.id)) {
      if (import.meta.env.MODE === 'development') {
        window.open(app.path, "_blank");
      }
    } else {
      navigate(app.path);
    }
  }
  
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
            onClick={handleNavigateCallback(app)}
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
