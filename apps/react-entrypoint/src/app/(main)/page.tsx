"use client";

import { Card, Center, Kbd, SimpleGrid, Stack, Text } from "@mantine/core";
import { useRouter } from "next/navigation";
import { APPS, type AppDefinition } from "@/shared/config/apps";

export default function Home() {
  const router = useRouter();

  const handleNavigateCallback = (app: AppDefinition) => () => {
    const backendServices = ["openapi", "graphql", "jeager", "admin"];

    if (backendServices.includes(app.id)) {
      if (process.env.NODE_ENV === "development") {
        window.open(app.path, "_blank");
      }
    } else {
      router.push(app.path);
    }
  };

  return (
    <Stack align="center" justify="center" h="100%" gap="xl">
      <Text size="xl" fw={700}>
        Applications
      </Text>

      <SimpleGrid cols={{ base: 1, sm: 2, md: 4 }} spacing="lg">
        {APPS.map((app) => (
          <Card
            key={app.id}
            shadow="sm"
            padding="sm"
            radius="md"
            withBorder
            style={{ cursor: "pointer", minWidth: 100 }}
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

      {/* <Button onClick={() => logMessage()}>Log Message</Button> */}
    </Stack>
  );
}
