import type { SpotlightActionData } from '@mantine/spotlight';
import { closeSpotlight } from '@mantine/spotlight';
import type { AppDefinition } from '../../../shared/config/apps';
import type { NavigateFunction } from 'react-router-dom';

export function buildSpotlightActions(
  apps: AppDefinition[],
  navigate: NavigateFunction,
): SpotlightActionData[] {
  return apps.map((app) => ({
    id: app.id,
    label: app.label,
    description: app.description,
    leftSection: app.icon,
    onClick: () => {
      navigate(app.path);
      closeSpotlight();
    },
  }));
}
