import { createTheme, MantineProvider } from "@mantine/core";
import type { ReactNode } from "react";

const theme = createTheme({
  primaryColor: "cyan",
  fontFamily: "system-ui, sans-serif",
});

interface ThemeProviderProps {
  children: ReactNode;
}

export function ThemeProvider({ children }: ThemeProviderProps) {
  return (
    <MantineProvider theme={theme} defaultColorScheme="auto">
      {children}
    </MantineProvider>
  );
}
