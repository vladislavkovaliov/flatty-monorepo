import "@mantine/core/styles.css";

import { Outlet, Link, useLocation } from "react-router-dom";

import {
  MantineProvider,
  createTheme,
  AppShell,
  Burger,
  NavLink,
  Group,
  Text,
  Box,
} from "@mantine/core";
import { useDisclosure } from "@mantine/hooks";
import { IconHome, IconSettings } from "@tabler/icons-react";

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
        <AppShell
          header={{ height: 60 }}
          navbar={{
            width: 250,
            breakpoint: "sm",
            collapsed: { mobile: !opened },
          }}
          padding={0}
        >
          <AppShell.Header>
            <Group h="100%" px="md">
              <Burger
                opened={opened}
                onClick={toggle}
                hiddenFrom="sm"
                size="sm"
              />
              <Text size="xl" fw={700}>
                React Wrapper
              </Text>
            </Group>
          </AppShell.Header>

          <AppShell.Navbar p="md">
            <Box mb="md">
              <Text fw={500} mb="xs">
                Menu
              </Text>
              {MENU_ITEMS.map((item) => (
                <NavLink
                  key={item.path}
                  label={item.label}
                  leftSection={item.icon}
                  active={isActiveLink(item.path)}
                  component={Link}
                  to={item.path}
                />
              ))}
            </Box>
          </AppShell.Navbar>

          <AppShell.Main
            style={{ height: "100%", display: "flex", flexDirection: "column" }}
          >
            <Outlet />
          </AppShell.Main>
        </AppShell>
      </SingleTabManagerWrapper>
    </MantineProvider>
  );
}
