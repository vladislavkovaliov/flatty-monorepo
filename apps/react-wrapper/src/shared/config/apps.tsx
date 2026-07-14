import type { ReactNode } from 'react';
import { IconHome, IconHome2, IconCategory2, IconCoin, IconSettings2, IconBrandGraphql, IconApi } from '@tabler/icons-react';

export interface AppDefinition  {
  id: string;
  label: string;
  description: string;
  path: string;
  icon: ReactNode;
}

export const APPS: AppDefinition[] = [
  {
    id: 'home',
    label: 'Home',
    description: 'Go back to home page',
    path: '/',
    icon: <IconHome size={24} stroke={1.5} />,
  },
  {
    id: 'resident',
    label: 'Resident',
    description: 'Manage residents',
    path: '/resident',
    icon: <IconHome2 size={24} stroke={1.5} />,
  },
  {
    id: 'settings',
    label: 'Settings',
    description: 'Manage settings',
    path: '/settings',
    icon: <IconSettings2 size={24} stroke={1.5} />,
  },
  {
    id: 'categories',
    label: 'Categories',
    description: 'Manage categories',
    path: '/categories',
    icon: <IconCategory2 size={24} stroke={1.5} />,
  },
  {
    id: 'expenses',
    label: 'Expenses',
    description: 'Manage expenses',
    path: '/expenses',
    icon: <IconCoin size={24} stroke={1.5} />,
  },
  {
    id: 'openapi',
    label: 'Open API',
    description: 'Open API',
    path: '/api/swagger/index.html',
    icon: <IconApi size={24} stroke={1.5} />,
  },
  {
    id: 'graphql',
    label: 'Open Graphql',
    description: 'Open Graphql',
    path: '/graphql',
    icon: <IconBrandGraphql size={24} stroke={1.5} />,
  },
];