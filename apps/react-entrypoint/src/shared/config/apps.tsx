import {
  IconApi,
  IconBrandGraphql,
  IconBrandMinecraft,
  IconCategory2,
  IconCoin,
  IconHome,
  IconHome2,
  IconReportMoney,
  IconScanTraces,
  IconSettings2,
  IconUsers,
} from "@tabler/icons-react";
import type { ReactNode } from "react";

export interface AppDefinition {
  id: string;
  label: string;
  description: string;
  path: string;
  icon: ReactNode;
}

export const APPS: AppDefinition[] = [
  {
    id: "home",
    label: "Home",
    description: "Go back to home page",
    path: "/",
    icon: <IconHome size={24} stroke={1.5} />,
  },
  {
    id: "resident",
    label: "Resident",
    description: "Manage residents",
    path: "/resident",
    icon: <IconHome2 size={24} stroke={1.5} />,
  },
  {
    id: "spending",
    label: "Spending",
    description: "View total and average spending",
    path: "/spending",
    icon: <IconReportMoney size={24} stroke={1.5} />,
  },
  {
    id: "settings",
    label: "Settings",
    description: "Manage settings",
    path: "/settings",
    icon: <IconSettings2 size={24} stroke={1.5} />,
  },
  {
    id: "categories",
    label: "Categories",
    description: "Manage categories",
    path: "/categories",
    icon: <IconCategory2 size={24} stroke={1.5} />,
  },
  {
    id: "expenses",
    label: "Expenses",
    description: "Manage expenses",
    path: "/expenses",
    icon: <IconCoin size={24} stroke={1.5} />,
  },
  {
    id: "users",
    label: "Users",
    description: "Manage users",
    path: "/users",
    icon: <IconUsers size={24} stroke={1.5} />,
  },
  {
    id: "openapi",
    label: "Open API",
    description: "Open API",
    path: "/api/swagger/index.html",
    icon: <IconApi size={24} stroke={1.5} />,
  },
  {
    id: "graphql",
    label: "Open Graphql",
    description: "Open Graphql",
    path: "/graphql",
    icon: <IconBrandGraphql size={24} stroke={1.5} />,
  },
  {
    id: "jeager",
    label: "Open Jaeger",
    description: "Open Jaeger",
    path: "http://localhost:16686",
    icon: <IconScanTraces size={24} stroke={1.5} />,
  },
  {
    id: "admin",
    label: "Open admin",
    description: "Open admin",
    path: "http://localhost:8080/admin/applications",
    icon: <IconBrandMinecraft size={24} stroke={1.5} />,
  },
];
