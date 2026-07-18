import {
  AppShell as MantineAppShell,
  Box as MantineBox,
  NavLink as MantinveNavLink,
} from "@mantine/core"

// TODO: need to be passed as props
import {  Link } from "react-router-dom";

interface IAppShellNavbarProps {
  isActiveFn: (path: string) => boolean
  items: Array<{
    label: string;
    path: string;
    icon: React.JSX.Element;
  }>;
}


export function AppShellNavbar({items, isActiveFn}: IAppShellNavbarProps) {
  return (
    <MantineAppShell.Navbar p="md">
      <MantineBox mb="md">
        {items.map((item) => (
          <MantinveNavLink
            key={item.path}
            label={item.label}
            leftSection={item.icon}
            active={isActiveFn(item.path)}
            component={Link}
            to={item.path}
          />
        ))}
      </MantineBox>
    </MantineAppShell.Navbar>
  )
}