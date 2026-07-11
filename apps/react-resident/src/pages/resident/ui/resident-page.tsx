import { Box, Container, Tabs, Title } from '@mantine/core';
import { ResidentTableTab } from './resident-table-tab';
import { ResidentCreateTab } from './resident-create-tab';

export function ResidentPage() {
  
  
  return (
    <>
      <Box py="sm">
        <Container fluid>
          <Box mb="sm">
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

        <Tabs.Panel value="resident-table" pt="md">
          <ResidentTableTab />
        </Tabs.Panel>

        <Tabs.Panel value="resident-create" pt="md">
          <ResidentCreateTab />
        </Tabs.Panel>
      
      </Tabs>
    </>
  );
}
