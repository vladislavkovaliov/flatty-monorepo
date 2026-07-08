# flatty-budget-monorepo

Nx monorepo with micro-frontends (React) + backends (Go/NestJS).

## Frontend

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


## Structure

```
apps/
├── react-launcher/     # Vite, shell-хост для MF (port 5175)
├── react-settings/     # Webpack UMD, remote MF (port 5174)
├── react-wrapper/      # Vite, shell-хост для MF (port 5174)
├── go-api/             # Go + Gin (port 8080)
├── nest-graphql/       # NestJS + auth guard + config (port 3000)
└── nginx-proxy/        # Nginx config
packages/
└── shared/             # @flatty-budget/shared — общие типы (IAppConfig, IAppComponent)
tools/
├── review/             # Go CLI — code review через Ollama
└── sql/                # Go CLI — поиск дубликатов SQL
```

## Команды

| Команда | Описание |
|---------|----------|
| `npm run dev` | Запустить всё параллельно |
| `npm run dev:frontend` | Только фронт (launcher + settings + wrapper) |
| `npm run dev:backend` | Только бек (go-api + nest-graphql) |
| `npm run build` | Собрать всё |
| `npm run graph` | Визуализация графа зависимостей |

## Микрофронты (custom script-injection)

- Remote-ы (react-settings) собираются как UMD: `window.ext-apps.settings`
- Shell-ы (react-launcher, react-wrapper) загружают их через `<script>` в runtime
- Типы: `IAppConfig`, `IAppComponent` — в `@flatty-budget/shared`

## Инфраструктура

```bash
docker compose up postgres redpanda rabbitmq
```

Приложения в dev запускаются на хосте через Nx. Docker — только для инфраструктуры.

## Interval tools

### tools/review

Go CLI for code review. See `tools/review/main.go` for details.

```bash
cd tools/review
go build -o review .
./review                  # review staged changes (default: shop-graphql-nestjs/)
./review --dir apps/      # review staged changes in apps/
./review --lang go        # review Go code (default: TypeScript)
./review --help           # flags: --dir, --lang, --model, --ollama-url
```

### tools/sql

Go CLI for finding duplicate SQL between Go and NestJS.

```bash
cd tools/sql
go build -o sql .
./sql                     # find duplicate SQL
./sql --help              # flags: --model, --ollama-url
```


## Documentation

Always prefer project documentation over assumptions.

Available documentation:

- docs/mantine.txt
- docs/tanstack.txt