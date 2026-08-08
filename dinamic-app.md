# Centralize the micro-frontend application registry in go-api

## Context

**How does `react-entrypoint` currently get its app list from `react-launcher`?** It doesn't — `apps/react-entrypoint/src/applications/*.ts` is a byte-for-byte duplicate of `apps/react-launcher/src/applications/*.ts`, but it's **never imported anywhere** in react-entrypoint (confirmed via grep). It's dead code. What react-entrypoint actually uses is a *third*, independent copy: `bundleName`/`cssBundleName`/`proxyBasePath`/`basePath` are hardcoded inline in `src/app/(main)/resident/[[...slug]]/page.tsx` and `.../settings/[[...slug]]/page.tsx`, and the remote origins/ports are hardcoded again in `next.config.mjs` rewrites. So today there are **three hand-synced copies** of the same registry (react-launcher's `applications/*.ts`, react-entrypoint's dead `applications/*.ts`, and react-entrypoint's inline hardcoding), and no shared source of truth. `react-entrypoint` also reimplemented its own local `IAppConfig`/`IAppComponent` types in `src/core/micro-frontend-host/domain.ts` instead of using the shared `@flatty-budget/mfe-host` package that `react-launcher` already depends on.

**Answering the four questions directly:**
1. **Best approach**: add a DB-backed `applications` table to go-api (per user decision) with one row per app per environment, exposed via a `GET /api/applications` endpoint (public/unauthenticated) that go-api filters to its own deployment environment, plus auth-protected `POST`/`PUT`/`DELETE` endpoints (per user decision) so rows can be changed without a redeploy. Both frontends fetch the list once via TanStack Query before rendering the micro-frontend host, instead of importing static files.
2. **Benefits**: one source of truth instead of three; new/changed MF URLs are a data change via an authenticated API call, not a multi-file code change across two apps; react-entrypoint's dead duplicate folder and reinvented types get deleted; react-launcher's `AppConfigurator` and react-entrypoint's proxy rewrites derive from the same data.
3. **DB storage**: yes, per your decision — table `applications` with `env`, `url` (remote origin), `bundle_js`, `style_url`, plus path fields, matching the existing `domains/repos/services/http` layering already used for `user_settings`, with full CRUD so the registry is editable at runtime.
4. **Different envs**: the `env` column lets one table hold rows for dev/staging/prod. go-api reads its own `APP_ENV` (new config field, same pattern as existing `PORT`/`DATABASE_URL` env vars) and the `GET /api/applications` endpoint only returns rows matching that env — no coordination needed from the frontends, and no query-param plumbing. Writes (`POST`/`PUT`) take `env` explicitly in the body, so one deployment could seed/update rows for another env ahead of a rollout if needed.

## 1. Database migration

New file `apps/go-api/migrations/006_add_applications.sql`, following the shape of `004_add_user_settings.sql`:

```sql
CREATE TABLE applications (
    id               BIGSERIAL   PRIMARY KEY,
    name             TEXT        NOT NULL,
    env              TEXT        NOT NULL,
    bundle_js        TEXT        NOT NULL,
    style_url        TEXT,
    remote_origin    TEXT        NOT NULL,
    proxy_base_path  TEXT        NOT NULL,
    base_path        TEXT        NOT NULL DEFAULT '/',
    created_at       TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at       TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (name, env)
);

CREATE TRIGGER update_applications_updated_at
    BEFORE UPDATE ON applications
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

INSERT INTO applications (name, env, bundle_js, style_url, remote_origin, proxy_base_path, base_path) VALUES
    ('app',      'development', 'app',      'styles', 'http://localhost:8080', '/external-app',      '/'),
    ('settings', 'development', 'settings', 'styles', 'http://localhost:8081', '/external-settings', '/'),
    ('resident', 'development', 'resident', 'styles', 'http://localhost:8082', '/external-resident', '/');
```

Append the new file to the hand-maintained slice in `apps/go-api/cmd/migrate/main.go` (same manual step required for `005_...` — this repo's migration runner is not automatic).

## 2. go-api backend (mirrors the existing `user_settings` / `category` layering)

- **`domains/applications/applications.go`**: `Application` struct (private fields + getters: `ID`, `Name`, `Env`, `BundleJS`, `StyleURL`, `RemoteOrigin`, `ProxyBasePath`, `BasePath`), constructed via `NewApplication(...)`.
- **`domains/applications/repository.go`**: `ApplicationsRepository` interface:
  ```go
  type ApplicationInput struct {
      Name, Env, BundleJS, StyleURL, RemoteOrigin, ProxyBasePath, BasePath string
  }
  type ApplicationsRepository interface {
      ListByEnv(ctx context.Context, env string) ([]*Application, error)
      Create(ctx context.Context, input ApplicationInput) (*Application, error)
      Update(ctx context.Context, id int64, input ApplicationInput) (*Application, error)
      Delete(ctx context.Context, id int64) error
  }
  ```
- **`repos/applications/applications.go`**: `NewPgxRepository(pool)` implementing the interface — `ListByEnv` via `SELECT ... WHERE env = $1 ORDER BY name`, `Create`/`Update` via `INSERT ... RETURNING` / `UPDATE ... WHERE id = $1 RETURNING`, `Delete` via `DELETE WHERE id = $1`, following pgx conventions already used in `repos/user_settings` and `repos/category`.
- **`services/applications/applications.go`**: thin `Service` wrapping the repo — `List(ctx, env)`, `Create(ctx, input)`, `Update(ctx, id, input)`, `Delete(ctx, id)`.
- **`http/dto/applications.go`**: `ApplicationResponse{ ID, Name, Env, BundleJS, StyleURL, RemoteOrigin, ProxyBasePath, BasePath }` for reads, `CreateApplicationRequest`/`UpdateApplicationRequest` (JSON + `binding:"required"` tags) for writes, matching the request/response DTO split already used in `http/dto/category.go`.
- **`http/handlers/applications.go`**: `ApplicationsHandler` with:
  - `List(c *gin.Context)` — no auth required (this is deployment config needed before the frontend can even reach a login-gated remote app), following the unauthenticated `ConfigHandler.Health` pattern. Resolves `env` from `cfg.AppEnv` (not a query param), maps domain → DTO.
  - `Create`, `Update`, `Delete` — bind the request DTO, call the service, return the updated resource / 204. These sit behind the same session `authMw` used by `/categories` and `/resident-location` (go-api has no separate admin role today, so this follows the existing "any authenticated user" convention rather than inventing one).
  All handlers use `context.WithTimeout(..., 5*time.Second)` like the existing handlers.
- **`internal/config/config.go`**: add `AppEnv string` field, loaded via `getEnv("APP_ENV", "development")`, same pattern as the other fields. Update `.env.example` with `APP_ENV=development`.
- **`cmd/api/wiring.go`**: add `wireApplications(rg *gin.RouterGroup, pool *pgxpool.Pool, cfg *config.Config, authMw gin.HandlerFunc)`:
  ```go
  func wireApplications(rg *gin.RouterGroup, pool *pgxpool.Pool, cfg *config.Config, authMw gin.HandlerFunc) {
      repo := applicationsrepo.NewPgxRepository(pool)
      svc := applicationsservice.New(repo)
      h := handlers.NewApplicationsHandler(svc, cfg)

      rg.GET("/applications", h.List) // public, needed before render

      protected := rg.Group("", authMw)
      protected.POST("/applications", h.Create)
      protected.PUT("/applications/:id", h.Update)
      protected.DELETE("/applications/:id", h.Delete)
  }
  ```
  Call it from `setupRouter` next to `wireConfig(api)`.

## 3. Shared frontend client (`packages/mfe-host`)

Add a new module, e.g. `packages/mfe-host/src/queries/applications.ts`:
- `ApplicationConfig` type mirroring the go-api DTO (`name`, `bundleJs`, `styleUrl`, `remoteOrigin`, `proxyBasePath`, `basePath`).
- `useApplications()` — a TanStack Query `useQuery` hook (per the repo's "queries for GET" convention) that `fetch`-es `/api/applications` and returns the typed list. Uses a relative URL so it rides the existing dev proxies (`vite.launcher.config.ts` already proxies `/api` → `localhost:8080`; `next.config.mjs` already proxies `/api/:path*` → `localhost:8080`) — **no CORS changes needed in go-api**.

Add `@tanstack/react-query` as a dependency of `packages/mfe-host` (it's already used elsewhere in the repo, e.g. `react-resident`, `react-settings`, `packages/sdk`, just not here yet).

## 4. `react-launcher` changes

- `apps/react-launcher/package.json` already depends on `@flatty-budget/mfe-host` — no new dependency needed.
- `src/AppConfigurator.tsx`: replace `import * as microApps from "./applications"` and the `appFactories`/`availableConfigs` static array with `useApplications()`. Feed the `Select`'s `data` from the query result instead of `availableConfigs.map(...)`; handle loading state (Mantine `Select` `disabled` while loading is enough, no need for a full skeleton).
- Delete `apps/react-launcher/src/applications/index.ts`, `react-app.ts`, `react-resident.ts`, `react-settings.ts` — fully replaced by the DB-backed list.

## 5. `react-entrypoint` changes

- Add `@flatty-budget/mfe-host` to `apps/react-entrypoint/package.json` dependencies.
- Delete the dead, unused `apps/react-entrypoint/src/applications/` folder (confirmed zero imports).
- In `src/core/micro-frontend-host/domain.ts`, drop the locally reimplemented `IAppConfig`/`IAppComponent` and import them from `@flatty-budget/mfe-host` instead, matching react-launcher.
- `resident/[[...slug]]/page.tsx` and `settings/[[...slug]]/page.tsx` are already `"use client"` components, so they can call `useApplications()` directly (no server-fetch plumbing needed). Look up the row by `name` (`"resident"` / `"settings"`) and pass `bundleName`, `cssBundleName`, `proxyBasePath`, `basePath` from the query result instead of the hardcoded literals. Show the existing `<Loader />` while the query is loading, same as the current `dynamic(..., { loading: () => <Loader /> })` behavior.
- `next.config.mjs`: make `rewrites()` fetch `GET http://localhost:8080/api/applications` once when the Next server starts (async `rewrites()` is supported by Next), and generate the `/external-<name>/:path*` → `<remoteOrigin>/:path*` entries from the DB rows instead of the three hardcoded destinations. Keep the `/api/:path*` → go-api and `/api/auth`, `/graphql` rewrites as-is (unrelated to this registry). Note: like all Next rewrite config, this is resolved once at server start, not per-request — a DB change still requires a restart to take effect, same cost as editing the file today.

## 6. SSR admin UI (inside go-api)

An htmx-powered, server-rendered admin page for managing `applications` rows directly, reusing the same `services/applications` layer as the JSON API — no new Go dependencies (stdlib `html/template` + `embed`), no CDN dependency (htmx vendored as a same-origin static file so it works under the existing `default-src 'self'` CSP from `secure.SecurityMiddleware`).

- **`services/applications`**: add `ListAll(ctx context.Context) ([]*Application, error)` (no env filter) alongside the existing `List(ctx, env)`, so the admin view can see and manage rows across *all* environments, not just the deployment's own `APP_ENV`. Repository gets a matching `ListAll` (`SELECT ... ORDER BY env, name`).
- **`http/templates/`** (new, embedded via `//go:embed templates/*.html` in a small `templates.go`):
  - `layout.html` — minimal page shell, pulls in `/static/htmx.min.js`.
  - `applications_list.html` — full admin page: a create form (name/env/bundle_js/style_url/remote_origin/proxy_base_path/base_path) that `hx-post`s to `/admin/applications`, plus `{{template "applications_table" .}}`.
  - `applications_table.html` — the `<table id="applications-table">` partial, one row per app grouped by `env`; each row has an "Edit" button (`hx-get="/admin/applications/{{.ID}}/edit"`, swaps the row for an inline edit form) and a "Delete" button (`hx-delete="/admin/applications/{{.ID}}"`, swaps the row away). This partial is reused as the response body after create/update/delete so htmx can swap it back in.
  - `applications_row_edit.html` — inline `<tr>` edit form partial, `hx-put`s back to `/admin/applications/{{.ID}}`.
- **`http/static/htmx.min.js`**: vendored file (checked into the repo, not fetched from a CDN at request time), served via `embed.FS` + `r.StaticFS("/static", http.FS(staticFS))`.
- **`http/handlers/admin_applications.go`**: `AdminApplicationsHandler` wrapping the same `*applicationsservice.Service`:
  - `ListPage(c)` → `svc.ListAll`, renders `applications_list.html`.
  - `CreateRow(c)` → binds form via `c.ShouldBind`, `svc.Create`, re-renders the `applications_table` partial.
  - `EditRowForm(c)` → renders `applications_row_edit.html` for one row.
  - `UpdateRow(c)` → `svc.Update`, renders the updated `applications_table` partial (simplest correct option — full-table refresh, not just the one row, avoids partial-vs-full template drift).
  - `DeleteRow(c)` → `svc.Delete`, renders the refreshed `applications_table` partial.
- **`cmd/api/main.go` / `setupRouter`**: load templates via `r.SetHTMLTemplate(template.Must(template.ParseFS(templatesFS, "templates/*.html")))` and mount static assets via `r.StaticFS(...)`.
- **`cmd/api/wiring.go`**: new `wireAdminApplications(r *gin.Engine, pool *pgxpool.Pool, authMw gin.HandlerFunc)`:
  ```go
  func wireAdminApplications(r *gin.Engine, pool *pgxpool.Pool, authMw gin.HandlerFunc) {
      repo := applicationsrepo.NewPgxRepository(pool)
      svc := applicationsservice.New(repo)
      h := handlers.NewAdminApplicationsHandler(svc)

      admin := r.Group("/admin")
      admin.Use(secure.SecurityMiddleware(), authMw)

      admin.GET("/applications", h.ListPage)
      admin.POST("/applications", h.CreateRow)
      admin.GET("/applications/:id/edit", h.EditRowForm)
      admin.PUT("/applications/:id", h.UpdateRow)
      admin.DELETE("/applications/:id", h.DeleteRow)
  }
  ```
  Called from `main()` (or `setupRouter`) alongside the `/api` wiring — note it's mounted on `r` directly (`/admin/...`), not on the `/api` router group, since this is a browser-facing page, not a JSON API surface.

  Note: `authMw` already reads the `better-auth.session_token` cookie (see `internal/auth/middleware.go`), so a logged-in browser session works here with no changes — but on 401 it currently returns raw JSON (`{"error":"unauthorized"}`) rather than redirecting to a login page. That's consistent with how every other protected route in go-api behaves today, so this plan leaves it as-is rather than introducing new redirect/login-page handling.

## Verification

1. `nx serve go-api` (with Postgres up via `docker compose up postgres`), run the new migration, then `curl http://localhost:8080/api/applications` and confirm it returns the three seeded `development` rows as JSON. Then, with a valid session cookie/token, exercise `POST`/`PUT`/`DELETE /api/applications` (e.g. update one row's `remote_origin`) and confirm `GET /api/applications` reflects the change, and that the write endpoints 401 without auth.
2. `nx serve react-launcher`: open the launcher UI, confirm the bundle `Select` is populated from the API (check Network tab for the `/api/applications` call) and picking one still fills the form correctly.
3. `nx serve react-entrypoint`: visit `/resident` and `/settings`, confirm the micro-frontends still load through the proxy rewrites, and check the terminal/server log at startup to confirm the dynamic rewrites were built from the API response (e.g. temporarily point a seeded row at a wrong port and confirm the proxy 404s/misroutes as expected).
4. Confirm `apps/react-entrypoint/src/applications/` and `apps/react-launcher/src/applications/` no longer exist, and `grep -r "core/micro-frontend-host/domain" apps/react-entrypoint/src` shows the local `IAppConfig`/`IAppComponent` definitions removed in favor of the `@flatty-budget/mfe-host` import.
5. With a logged-in browser session, visit `http://localhost:8080/admin/applications`: confirm the table lists all seeded rows across envs, creating a row via the form appends it without a full page reload (htmx swap), clicking "Edit" swaps in an inline form, submitting it updates the row, and "Delete" removes it — then re-check `GET /api/applications` reflects the change. Also confirm an unauthenticated request to `/admin/applications` is rejected (401), and that no browser console CSP violations occur (htmx served from `/static`, not a CDN).
