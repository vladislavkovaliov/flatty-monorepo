import "./index.css";
import { useState } from "react";
import {
  MantineProvider,
  Drawer,
  ActionIcon,
  Stack,
  Text,
  Code,
  Group,
} from "@mantine/core";
import { AppConfigurator } from "./AppConfigurator";
import type { IAppConfig } from "./types/external-app-config.type";
import "@mantine/core/styles.css";
import { MicrofrontendHost } from "./utils/MicrofrontendHost";
import { IconSettings, IconRestore } from "@tabler/icons-react";

export function Launcher() {
  const [config, setConfig] = useState<IAppConfig | null>(() => {
    try {
      const _config = localStorage.getItem("react-applications");

      if (_config !== null) {
        return JSON.parse(_config);
      } else {
        return null;
      }
    } catch (error) {
      console.log("Failed to load config from local storage", error);
    }
  });
  const [drawerOpened, setDrawerOpened] = useState(false);

  const handleStart = (cfg: IAppConfig) => {
    setConfig(cfg);
  };

  const handleRestoreCallback = () => {
    localStorage.removeItem("react-applications");
    window.location.reload();
  };

  const renderConfigContent = () => {
    if (!config) {
      return <Text c="dimmed">No configuration set</Text>;
    }

    return (
      <Stack gap="xs">
        <ConfigItem label="Bundle" value={config.bundleName} />
        <ConfigItem label="CSS Bundle" value={config.cssBundleName} />
        <ConfigItem label="Remote Origin" value={config.remoteOrigin} />
        <ConfigItem label="Environment" value={config.env} />
        <ConfigItem label="Host Type" value={config.hostType} />
        <ConfigItem
          label="Feature Flags"
          value={JSON.stringify(config.featureFlags, null, 2)}
          isCode
        />
      </Stack>
    );
  };

  return (
    <MantineProvider>
      {config && config.bundleName ? (
        <MicrofrontendHost
          bundleName={config.bundleName}
          cssBundleName={config.cssBundleName || "styles"}
          remoteOrigin={config.remoteOrigin}
          basePath={config.basePath}
          proxyBasePath={config.proxyBasePath}
          config={{
            env: config.env,
            featureFlags: config.featureFlags,
            hostType: config.hostType,
          }}
        />
      ) : (
        <AppConfigurator onStart={handleStart} />
      )}

      <Group
        style={{
          position: "fixed",
          bottom: 20,
          right: 20,
          zIndex: 1000,
        }}
      >
        <ActionIcon
          size="lg"
          radius="xl"
          variant="filled"
          onClick={handleRestoreCallback}
          disabled={config === null}
        >
          <IconRestore size={20} />
        </ActionIcon>
        <ActionIcon
          size="lg"
          radius="xl"
          variant="filled"
          onClick={() => setDrawerOpened(true)}
        >
          <IconSettings size={20} />
        </ActionIcon>
      </Group>

      <Drawer
        opened={drawerOpened}
        onClose={() => setDrawerOpened(false)}
        position="right"
        size="sm"
        title="Current Configuration"
      >
        {renderConfigContent()}
      </Drawer>
    </MantineProvider>
  );
}

function ConfigItem({
  label,
  value,
  isCode,
}: {
  label: string;
  value?: string;
  isCode?: boolean;
}) {
  if (!value) return null;

  return (
    <div>
      <Text size="sm" fw={500} c="dimmed">
        {label}
      </Text>
      {isCode ? <Code block>{value}</Code> : <Text>{value}</Text>}
    </div>
  );
}
