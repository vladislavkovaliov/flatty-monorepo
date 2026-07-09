import "@mantine/core/styles.css";

import { Outlet, Link, useLocation } from "react-router-dom";

import {
  MantineProvider,
  createTheme,
  AppShell as MantineAppShell, 
} from "@mantine/core";
import { useDisclosure } from "@mantine/hooks";
import { IconHome, IconSettings } from "@tabler/icons-react";
import { AppShellHeader } from "./components/app-shell-header"
import { AppShellNavbar } from "./components/app-shell-navbar"

import { SingleTabManagerWrapper } from "./components/single-tab-manager";

const theme = createTheme({
  primaryColor: "cyan",
  fontFamily: "system-ui, sans-serif",
});

const MENU_ITEMS = [
  { label: "Home", path: "/", icon: <IconHome size={16} stroke={1.5} /> },
  { label: "Settings", path: "/settings", icon: <IconSettings size={16} stroke={1.5} /> },
];

export default function App() {
  const [opened, { toggle }] = useDisclosure(true);
  const location = useLocation();

  const isActiveLink = (path: string) => location.pathname === path;
  
  return (
    <MantineProvider theme={theme} defaultColorScheme="dark">
      <SingleTabManagerWrapper>
        <MantineAppShell
          header={{ height: 60 }}
          navbar={{
            width: 250,
            breakpoint: "sm",
            collapsed: { mobile: !opened },
          }}
          padding={0}
        >
          <AppShellHeader opened={opened} toggle={toggle} />
          <AppShellNavbar items={MENU_ITEMS} isActiveFn={isActiveLink} />
          <MantineAppShell.Main
            style={{ height: "100%", display: "flex", flexDirection: "column" }}
          >
            <Outlet />
          </MantineAppShell.Main>
        </MantineAppShell>
      </SingleTabManagerWrapper>
    </MantineProvider>
  );
}
