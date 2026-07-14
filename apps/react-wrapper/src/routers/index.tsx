import { createBrowserRouter } from "react-router-dom";
import { HomePage } from "../pages/home";
import { CategoriesPage } from "../pages/categories";
import App from "../App";

import * as microApps from "../applications";
import { MicrofrontendHost } from "../core/micro-frontend-host";
import { CreateCategoryPage } from "../pages/create-category";
import { ExpensesPage, CreateExpensePage } from "../pages/expenses";

type MicroAppConfig = {
  bundleName: string;
  cssBundleName?: string;
  remoteOrigin?: string;
  basePath?: string;
  proxyBasePath?: string;
};

type MicroAppFactory = () => MicroAppConfig;

type AppFactories = typeof microApps;
type AppConfigUnion = ReturnType<AppFactories[keyof AppFactories]>;
type BundleName = AppConfigUnion["bundleName"];
type AvailableConfigs = {
  [K in BundleName]: Extract<AppConfigUnion, { bundleName: K }>;
};

const availableConfigs = Object.fromEntries(
  Object.values(microApps).map((factory) => {
    const cfg = (factory as MicroAppFactory)();
    return [cfg.bundleName, cfg] as const;
  }),
) as AvailableConfigs;

export const router = createBrowserRouter([
  {
    path: "/",
    element: <App />,
    children: [
      {
        path: "/",
        element: <HomePage />,
      },
      {
        path: "/resident/*",
        element: (
          <MicrofrontendHost {...availableConfigs.resident} />
        ),
      },
      {
        path: "/categories",
        element: (
          <CategoriesPage />
        ),
      },
      {
        path: "/categories/create",
        element: (
          <CreateCategoryPage />
        ),
      },
      {
        path: "/expenses",
        element: (
          <ExpensesPage />
        ),
      },
      {
        path: "/expenses/create",
        element: (
          <CreateExpensePage />
        ),
      },
    ],
  },
  
]);
