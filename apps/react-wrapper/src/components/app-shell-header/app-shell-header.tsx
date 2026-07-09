import {
  AppShell as MantineAppShell,
  Text as MantineText,
  Group,
  Burger,
} from "@mantine/core"

interface IAppShellHeaderProps {
  opened: boolean;
  toggle: () => void;
}

export function AppShellHeader({opened, toggle}: IAppShellHeaderProps) {
  return (
    <MantineAppShell.Header>
      <Group h="100%" px="md">
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
    </MantineAppShell.Header>
  )
}