import { Tabs, Title, Container } from '@mantine/core';

import { LocalizationTab } from './localization-tab';
import { ThemesTab } from './themes-tab';
import { TimezoneTab } from './timezone-tab';

export function UserSettingsPage() {
  return (
    <Container fluid py="xl">
      <Title order={2} mb="lg">
        User Settings
      </Title>

      <Tabs defaultValue="localization" variant="default">
        <Tabs.List>
          <Tabs.Tab value="localization">Localization</Tabs.Tab>
          <Tabs.Tab value="themes">Themes</Tabs.Tab>
          <Tabs.Tab value="timezone">Timezone</Tabs.Tab>
        </Tabs.List>

        <Tabs.Panel value="localization" pt="lg">
          <LocalizationTab />
        </Tabs.Panel>

        <Tabs.Panel value="themes" pt="lg">
          <ThemesTab />
        </Tabs.Panel>

        <Tabs.Panel value="timezone" pt="lg">
          <TimezoneTab />
        </Tabs.Panel>
      </Tabs>
    </Container>
  );
}
