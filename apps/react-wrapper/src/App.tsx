import "@mantine/core/styles.css";
import "@mantine/dates/styles.css";
import "@mantine/charts/styles.css";
import "@mantine/spotlight/styles.css";

import { Outlet, useNavigate } from "react-router-dom";

import {
  AppShell,
  MantineProvider,
  createTheme,
} from "@mantine/core";
import { useDisclosure } from "@mantine/hooks";
import { Spotlight } from "@mantine/spotlight";

import { SingleTabManagerWrapper } from "./components/single-tab-manager";
import { AppShellHeader } from "./components/app-shell-header/app-shell-header";
import { APPS } from "./shared/config/apps";
import { buildSpotlightActions } from "./features/spotlight";

const theme = createTheme({
  primaryColor: "cyan",
  fontFamily: "system-ui, sans-serif",
});

export default function App() {
  const navigate = useNavigate();
  const [opened, { toggle }] = useDisclosure();

  const spotlightActions = buildSpotlightActions(APPS, navigate);

  return (
    <MantineProvider theme={theme} defaultColorScheme="dark">
      <AppShell header={{ height: 60 }}>
        <AppShellHeader opened={opened} toggle={toggle} />

        <AppShell.Main>
          <SingleTabManagerWrapper>
            <Outlet />
          </SingleTabManagerWrapper>
        </AppShell.Main>
      </AppShell>

      <Spotlight
        shortcut="mod + K"
        actions={spotlightActions}
        nothingFound="Nothing found..."
        highlightQuery
        searchProps={{
          placeholder: "Search applications...",
        }}
      />
    </MantineProvider>
  );
}
