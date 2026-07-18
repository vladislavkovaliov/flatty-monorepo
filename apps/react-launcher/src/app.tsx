import { Outlet } from "react-router-dom";
import {
  SingleTabManager,
  type SingleTabManagerOptions,
} from "single-active-browser-tab";

import {
  MantineProvider,
  createTheme,
} from "@mantine/core";
import { useEffect, useRef, useState } from "react";

import "@mantine/core/styles.css";

const theme = createTheme({
  primaryColor: "cyan",
  fontFamily: "system-ui, sans-serif",
});

export function App() {
  const [isActive, setIsActive] = useState<boolean | null>(null);

  const managerRef = useRef<SingleTabManager | null>(null);

  useEffect(() => {
    const singleTabManager = new SingleTabManager("broadcast", {
      onActive: () => {
        setIsActive(true);
      },
      onBlocked: () => {
        setIsActive(false);
      },
      logLevel: "log",
    } satisfies SingleTabManagerOptions);

    managerRef.current = singleTabManager;

    singleTabManager.start();

    return () => {
      singleTabManager.stop();
      managerRef.current = null;
    };
  }, []);

  const handleReloadCallback = () => {
    managerRef.current?.takeover();
  };

  return (
    <MantineProvider theme={theme} defaultColorScheme="dark">
      {isActive && <Outlet />}
      {isActive === false && (
        <div>
          <p>Application is already opened in other tabs.</p>
          <button onClick={handleReloadCallback}>Reload</button>
        </div>
      )}
    </MantineProvider>
  );
}
