import { Box, Container, Tabs, Title } from '@mantine/core';
import { ResidentTableTab } from './resident-table-tab';
import { ResidentCreateTab } from './resident-create-tab';

export function ResidentPage() {
  
  
  return (
    <>
      <Box py="xl">
        <Container fluid>
          <Box mb="lg">
            <Title order={2}>
              Resident
            </Title>
          </Box>
        </Container>
      </Box>

      <Tabs defaultValue="resident-table" variant="default">
        <Tabs.List>
          <Tabs.Tab value="resident-table">Table</Tabs.Tab>
          <Tabs.Tab value="resident-create">Create</Tabs.Tab>
        </Tabs.List>

        <Tabs.Panel value="resident-table" pt="lg">
          <ResidentTableTab />
        </Tabs.Panel>

        <Tabs.Panel value="resident-create" pt="lg">
          <ResidentCreateTab />
        </Tabs.Panel>
      
      </Tabs>
    </>
  );
}
