import { createBrowserRouter } from "react-router-dom";

import App from "#/App";

import { AuthGuard } from "#/features/auth/guard/auth-guard";
import { lazyLoad } from "#/shared/ui/lazy-load"
import { availableConfigs } from "./constants";


const ExpensesPage = lazyLoad(() =>
  import('#/pages/expenses').then(m => ({ default: m.ExpensesPage }))
);

const CreateExpensePage = lazyLoad(() =>
  import('#/pages/expenses').then(m => ({ default: m.CreateExpensePage }))
);

const UsersPage = lazyLoad(() =>
  import('#/pages/users').then(m => ({ default: m.UsersPage }))
);

const CreateCategoryPage = lazyLoad(() =>
  import('#/pages/create-category').then(m => ({ default: m.CreateCategoryPage }))
);

const CategoriesPage = lazyLoad(() =>
  import('#/pages/categories').then(m => ({ default: m.CategoriesPage }))
);

const HomePage = lazyLoad(() =>
  import('#/pages/home').then(m => ({ default: m.HomePage }))
);

const MicrofrontendHost = lazyLoad(() =>
  import('#/core/micro-frontend-host').then(m => ({ default: m.MicrofrontendHost }))
);

const LoginPage = lazyLoad(() =>
  import('#/pages/auth').then(m => ({ default: m.LoginPage }))
);

const RegisterPage = lazyLoad(() =>
  import('#/pages/auth').then(m => ({ default: m.RegisterPage }))
);

const AcceptInvitePage = lazyLoad(() =>
  import('#/pages/accept-invite').then(m => ({ default: m.AcceptInvitePage }))
);

const SetPasswordPage = lazyLoad(() =>
  import('#/pages/set-password').then(m => ({ default: m.SetPasswordPage }))
);

export const router = createBrowserRouter([
  {
    path: "/login",
    element: <LoginPage />,
  },
  {
    path: "/register",
    element: <RegisterPage />,
  },
  {
    path: "/accept-invite",
    element: <AcceptInvitePage />,
  },
  {
    path: "/set-password",
    element: <SetPasswordPage />,
  },
  {
    path: "/",
    element: <AuthGuard><App /></AuthGuard>,
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
        path: "/settings/*",
        element: (
          <MicrofrontendHost {...availableConfigs.settings} />
        ),
      },
      {
        path: "/users",
        element: (
          <UsersPage />
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
