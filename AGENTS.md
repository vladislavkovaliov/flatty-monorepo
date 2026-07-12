# flatty-budget-monorepo

Nx monorepo with micro-frontends (React) + backends (Go/NestJS).

## Projects

| App | Role | Stack | Port | AGENT.md |
|-----|------|-------|------|----------|
| react-launcher | Shell host (Vite) | React, Mantine v9 | 9000 | [→](apps/react-launcher/AGENTS.md) |
| react-wrapper | Shell host (Vite) | React | 5174 | [→](apps/react-wrapper/AGENTS.md) |
| react-settings | Remote MF (Webpack) | React | 8081 | [→](apps/react-settings/AGENTS.md) |
| react-resident | Remote MF (Webpack) | React, Mantine v8 | 8082 | [→](apps/react-resident/AGENTS.md) |
| go-api | REST API | Go + Gin | 8080 | [→](apps/go-api/AGENTS.md) |
| nest-graphql | GraphQL API | NestJS | 3000 | [→](apps/nest-graphql/AGENTS.md) |
| nginx-proxy | Reverse proxy | Nginx | 80 | [→](apps/nginx-proxy/AGENTS.md) |

## Tools

| Tool | Purpose | AGENT.md |
|------|---------|----------|
| review | Code review via Ollama | [→](tools/review/AGENTS.md) |
| sql | Find duplicate SQL | [→](tools/sql/AGENTS.md) |

## Shared Frontend Conventions

Apply to all React apps (react-launcher, react-wrapper, react-settings, react-resident).

Use:
- React
- TypeScript
- Mantine v8
- TanStack Query

Do not use:
- Material UI
- Ant Design
- Chakra UI

Prefer Mantine components before writing custom ones.

Prefer:
- Stack
- Group
- Flex
- SimpleGrid

Avoid custom CSS when Mantine provides a component.

Always use documented Mantine APIs.

Documentation:

docs/mantine.txt

### Code Style

- TypeScript strict
- Named exports
- Functional React components
- Small files
- Small functions

Avoid:
- any
- default exports
- magic strings
- deep nesting

### Data Fetching

Use TanStack Query.

Rules:
- queries for GET
- mutations for POST/PUT/PATCH/DELETE
- never fetch directly inside components unless explicitly requested

Documentation:

docs/tanstack.txt

## Commands

| Command | Description |
|---------|-------------|
| `npm run dev` | Запустить всё параллельно |
| `npm run dev:frontend` | Только фронт (launcher + settings + wrapper) |
| `npm run dev:backend` | Только бек (go-api + nest-graphql) |
| `npm run build` | Собрать всё |
| `npm run graph` | Визуализация графа зависимостей |

## Микрофронты (custom script-injection)

- Remote-ы (react-settings, react-resident) собираются как UMD: `window.ext-apps.settings`, `window.ext-apps.resident`
- Shell-ы (react-launcher, react-wrapper) загружают их через `<script>` в runtime
- Типы: `IAppConfig`, `IAppComponent` — в `@flatty-budget/shared`

## Инфраструктура

```bash
docker compose up postgres redpanda rabbitmq
```

Приложения в dev запускаются на хосте через Nx. Docker — только для инфраструктуры.

## Documentation

Always prefer project documentation over assumptions.

Available documentation:

- docs/mantine.txt
- docs/tanstack.txt
