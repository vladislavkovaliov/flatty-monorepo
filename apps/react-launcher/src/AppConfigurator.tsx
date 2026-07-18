import { useForm } from "@mantine/form";
import {
  Button,
  Group,
  Select,
  Stack,
  Textarea,
  TextInput,
  Title,
  Box,
} from "@mantine/core";

import type { IAppConfig } from "./types/external-app-config.type";
import * as microApps from "./applications";

interface ConfiguratorProps {
  onStart: (config: IAppConfig) => void;
}

type AppConfigFormValues = Omit<IAppConfig, "featureFlags"> & {
  featureFlags: string;
  bundleName?: string;
  cssBundleName?: string;
  remoteOrigin?: string;
  basePath?: string;
  proxyBasePath?: string;
};

type MicroAppConfig = {
  bundleName: string;
  cssBundleName?: string;
  remoteOrigin?: string;
  basePath?: string;
  proxyBasePath?: string;
};

type MicroAppFactory = () => MicroAppConfig;

const appFactories = microApps as Record<string, MicroAppFactory>;

const availableConfigs: MicroAppConfig[] = Object.keys(appFactories).map(
    (key: string) => appFactories[key]()
);

export function AppConfigurator({ onStart }: ConfiguratorProps) {
  const form = useForm<AppConfigFormValues>({
    initialValues: {
      env: "development",
      featureFlags: "{}",
      hostType: "react",
      basePath: "/",
      proxyBasePath: "/",
    },
  });

  const handleBundleChange = (value: string | null) => {
    if (!value) return;

    const selected = availableConfigs.find(
        (c: MicroAppConfig) => c.bundleName === value
    );

    if (selected) {
      form.setValues({
        bundleName: selected.bundleName,
        cssBundleName: selected.cssBundleName,
        remoteOrigin: selected.remoteOrigin,
        basePath: selected.basePath,
        proxyBasePath: selected.proxyBasePath,
      });
    }
  };

  const handleSubmit = (values: AppConfigFormValues) => {
    let featureFlags: Record<string, unknown> = {};

    try {
      featureFlags = JSON.parse(values.featureFlags || "{}");
    } catch {
      form.setFieldError("featureFlags", "invalid JSON");
      return;
    }

    const _config = {
      bundleName: values.bundleName,
      cssBundleName: values.cssBundleName,
      remoteOrigin: values.remoteOrigin,
      proxyBasePath: values.proxyBasePath,
      env: values.env,
      featureFlags,
      hostType: values.hostType,
      basePath: values.basePath,
    };

    localStorage.setItem('react-applications', JSON.stringify(_config));

    onStart(_config);
  };
  return (
      <Box maw={500} mx="auto" mt="xl">
        <Title order={2} mb="md">
          App Configuration
        </Title>
        <form
          onSubmit={form.onSubmit(handleSubmit)}
        >
          <Stack>
            <Select
              label="Bundle"
              placeholder="Select bundle"
              data={availableConfigs.map((c: MicroAppConfig) => ({
                value: c.bundleName,
                label: c.bundleName,
              }))}
              value={form.values.bundleName}
              onChange={handleBundleChange}
            />
            <TextInput
              label="CSS Bundle"
              placeholder="styles"
              {...form.getInputProps("cssBundleName")}
            />
            <TextInput
              label="Remote Origin"
              placeholder="http://localhost:8080"
              {...form.getInputProps("remoteOrigin")}
            />
            <TextInput
                label="Proxy"
                placeholder="/"
                {...form.getInputProps("proxyBasePath")}
            />
            <TextInput
                label="Basepath"
                placeholder="/"
                {...form.getInputProps("basePath")}
            />
            <TextInput
              label="Environment"
              placeholder="development"
              {...form.getInputProps("env")}
            />
            <Select
              label="Host Type"
              data={[
                { value: "react", label: "React" },
                { value: "angular", label: "Angular" },
                { value: "other", label: "Other" },
              ]}
              {...form.getInputProps("hostType")}
            />
            <Textarea
              label="Feature Flags (JSON)"
              autosize
              minRows={3}
              {...form.getInputProps("featureFlags")}
            />
            <Group mt="md">
              <Button type="submit">Start</Button>
            </Group>
          </Stack>
        </form>
      </Box>
  );
}
