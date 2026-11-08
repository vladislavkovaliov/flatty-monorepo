# flatty-budget-monorepo

Nx monorepo with micro-frontends (React) + backends (Go/NestJS).

## Projects

| App | Role | Stack | Port | AGENT.md |
|-----|------|-------|------|----------|
| react-launcher | Shell host (Vite) | React, Mantine v9 | 9000 | [→](apps/react-launcher/AGENTS.md) |
| react-settings | Remote MF (Webpack) | React | 8081 | [→](apps/react-settings/AGENTS.md) |
| react-resident | Remote MF (Webpack) | React, Mantine v8 | 8082 | [→](apps/react-resident/AGENTS.md) |
| go-api | REST API | Go + Gin | 8080 | [→](apps/go-api/AGENTS.md) |
| nest-graphql | GraphQL API | NestJS | 3000 | [→](apps/nest-graphql/AGENTS.md) |
| nginx-proxy | Reverse proxy | Nginx | 80 | [→](apps/nginx-proxy/AGENTS.md) |
| react-entrypoint | Shell host (Vite) | Nestjs | 3000 (3001) | [→](apps/react-entrypoint/AGENTS.md) |
| python-torrent-bot | Telegram-бот торрент-загрузок | Python, aiogram, FastAPI | 8090 | [→](apps/python-torrent-bot/AGENTS.md) |

## Tools

| Tool | Purpose | AGENT.md |
|------|---------|----------|
| review | Code review via Ollama | [→](tools/review/AGENTS.md) |
| sql | Find duplicate SQL | [→](tools/sql/AGENTS.md) |

## Shared Frontend Conventions

Apply to all React apps (react-launcher, react-wrapper, react-entrypoint, react-settings, react-resident).

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
| `npm run dev` | Run everything in parallel |
| `npm run dev:frontend` | Frontend only (launcher + settings + wrapper) |
| `npm run dev:backend` | Backend only (go-api + nest-graphql) |
| `npm run build` | Build everything |
| `npm run graph` | Visualize dependency graph |

## Micro-frontends (custom script-injection)

- Remotes (react-settings, react-resident) are built as UMD: `window.ext-apps.settings`, `window.ext-apps.resident`
- Shells (react-launcher, react-wrapper) load them via `<script>` at runtime
- Types: `IAppConfig`, `IAppComponent` — in `@flatty-budget/shared`

## Infrastructure

```bash
podman-compose -f docker-compose.yml --env-file .env.dev up -d postgres redpanda rabbitmq
```

Apps in dev run on host via Nx. Docker is for infrastructure only.

## Documentation

Always prefer project documentation over assumptions.

Available documentation:

- docs/mantine.txt
- docs/tanstack.txt

## graphify

This project has a knowledge graph at graphify-out/ with god nodes, community structure, and cross-file relationships.

When the user types `/graphify`, use the installed graphify skill or instructions before doing anything else.

Rules:
- When graphify-out/graph.json exists, run `graphify query "<question>"` before your first grep, search or source-file read for a codebase question, and put this rule in every subagent prompt that explores code. Read raw files only for lines the graph or an earlier read already located. Use `graphify path "<A>" "<B>"` for relationships and `graphify explain "<concept>"` for focused concepts. These return a scoped subgraph, usually much smaller than GRAPH_REPORT.md or raw grep output.
- Dirty graphify-out/ files are expected after hooks or incremental updates; dirty graph files are not a reason to skip graphify. Only skip graphify if the task is about stale or incorrect graph output, or the user explicitly says not to use it.
- If graphify-out/wiki/index.md exists, use it for broad navigation instead of raw source browsing.
- Read graphify-out/GRAPH_REPORT.md only for broad architecture review or when query/path/explain do not surface enough context.
- After modifying code, run `graphify update .` to keep the graph current (AST-only, no API cost).
