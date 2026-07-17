import "@mantine/core/styles.css";
import "@mantine/dates/styles.css";
import "@mantine/charts/styles.css";
import "@mantine/spotlight/styles.css";

import { Outlet, useNavigate } from "react-router-dom";

import {
  MantineProvider,
  createTheme,
} from "@mantine/core";
import { Spotlight } from "@mantine/spotlight";

import { SingleTabManagerWrapper } from "./components/single-tab-manager";
import { APPS } from "./shared/config/apps";
import { buildSpotlightActions } from "./features/spotlight";

const theme = createTheme({
  primaryColor: "cyan",
  fontFamily: "system-ui, sans-serif",
});

export default function App() {
  const navigate = useNavigate();

  const spotlightActions = buildSpotlightActions(APPS, navigate);

  return (
    <MantineProvider theme={theme} defaultColorScheme="dark">
      <SingleTabManagerWrapper>
        <Outlet />
      </SingleTabManagerWrapper>

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
