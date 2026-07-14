import * as microApps from "#/applications";
import type { AvailableConfigs, MicroAppFactory } from "./types";


export const availableConfigs = Object.fromEntries(
  Object.values(microApps).map((factory) => {
    const cfg = (factory as MicroAppFactory)();
    return [cfg.bundleName, cfg] as const;
  }),
) as AvailableConfigs;
