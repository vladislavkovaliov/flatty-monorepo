import { useState } from "react";
import { useNavigate } from "react-router-dom";

import {
  AppShell as MantineAppShell,
  Text as MantineText,
  Button,
  Group,
  Burger,
  Drawer,
  Stack,
} from "@mantine/core";

import { signOut, useSession } from "#/lib/auth-client";

function LogoutButton({ loading, onHandleSignOut}: {loading: boolean, onHandleSignOut: () => Promise<void>}) {
    return (
      <Button
        variant="outline"
        color="red"
        size="xs"
        loading={loading}
        onClick={onHandleSignOut}
      >
        Sign Out
      </Button>
    )
  }

interface IAppShellHeaderProps {
  opened: boolean;
  toggle: () => void;
}

export function AppShellHeader({ opened, toggle }: IAppShellHeaderProps) {
  const { data: session } = useSession();
  const navigate = useNavigate();
  const [signingOut, setSigningOut] = useState(false);

  const handleSignOut = async () => {
    setSigningOut(true);
    try {
      await signOut();
      navigate("/login");
    } catch {
      setSigningOut(false);
    }
  };

  return (
    <MantineAppShell.Header>
      <Group h="100%" px="md" justify="space-between">
        <Group>
          <Burger
            opened={opened}
            onClick={toggle}
            hiddenFrom="sm"
            size="sm"
          />
          <MantineText size="xl" fw={700}>
            React Wrapper
          </MantineText>
        </Group>

        <Group visibleFrom="sm">
          <MantineText size="sm" c="dimmed">
            {session?.user?.email}
          </MantineText>
          <LogoutButton  
            loading={signingOut}
            onHandleSignOut={handleSignOut} 
          />
        </Group>
      </Group>
      <Drawer
        opened={opened}
        onClose={toggle}
        size="100%"
        padding="md"
        hiddenFrom="sm"
        title="Menu"
      >
        <Stack gap="xs">
          <Stack>
            <MantineText size="sm" c="dimmed">
              {session?.user?.email}
            </MantineText>
            <LogoutButton  
              loading={signingOut}
              onHandleSignOut={handleSignOut} 
            />
          </Stack>
        </Stack>
      </Drawer>
    </MantineAppShell.Header>
  );
}