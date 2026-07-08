# flatty-budget-monorepo

Nx monorepo with micro-frontends (React/Angular) + backends (Go/NestJS).

## Структура

```
apps/
├── react-launcher/     # Vite, shell-хост для MF (port 5175)
├── react-settings/     # Webpack UMD, remote MF (port 5174)
├── angular-shell/      # Angular 21 SSR, shell-хост (port 4200)
├── go-api/             # Go + Gin (port 8080)
├── nest-graphql/       # NestJS + auth guard + config (port 3000)
└── nginx-proxy/        # Nginx config
packages/
└── shared/             # @flatty-budget/shared — общие типы (IAppConfig, IAppComponent)
tools/                  # Пока пусто
```

## Команды

| Команда | Описание |
|---------|----------|
| `npm run dev` | Запустить всё параллельно |
| `npm run dev:frontend` | Только фронт (launcher + settings + angular) |
| `npm run dev:backend` | Только бек (go-api + nest-graphql) |
| `npm run build` | Собрать всё |
| `npm run graph` | Визуализация графа зависимостей |

## Микрофронты (custom script-injection)

- Remote-ы (react-settings) собираются как UMD: `window.ext-apps.settings`
- Shell-ы (react-launcher, angular-shell) загружают их через `<script>` в runtime
- Типы: `IAppConfig`, `IAppComponent` — в `@flatty-budget/shared`

## Инфраструктура

```bash
docker compose up postgres redpanda rabbitmq
```

Приложения в dev запускаются на хосте через Nx. Docker — только для инфраструктуры.

## tools/review

Go CLI for code review. See `tools/review/main.go` for details.

```bash
cd tools/review
go build -o review .
./review                  # review staged changes (default: shop-graphql-nestjs/)
./review --dir apps/      # review staged changes in apps/
./review --lang go        # review Go code (default: TypeScript)
./review --help           # flags: --dir, --lang, --model, --ollama-url
```

## tools/sql

Go CLI for finding duplicate SQL between Go and NestJS.

```bash
cd tools/sql
go build -o sql .
./sql                     # find duplicate SQL
./sql --help              # flags: --model, --ollama-url
```