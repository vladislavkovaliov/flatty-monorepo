import type { SpotlightActionData } from "@mantine/spotlight";
import { closeSpotlight } from "@mantine/spotlight";
import type { AppDefinition } from "@/shared/config/apps";

export function buildSpotlightActions(
  apps: AppDefinition[],
  navigate: (app: AppDefinition) => void,
): SpotlightActionData[] {
  return apps.map((app) => ({
    id: app.id,
    label: app.label,
    description: app.description,
    leftSection: app.icon,
    onClick: () => {
      const backendServices = ["openapi", "graphql", "jaeger", "admin"];

      if (backendServices.includes(app.id)) {
        window.open(app.path, "_blank");
      } else {
        navigate(app);
        closeSpotlight();
      }
    },
  }));
}
