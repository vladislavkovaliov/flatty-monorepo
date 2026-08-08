# flatty-budget-monorepo

Nx monorepo with micro-frontends (React) + backends (Go/NestJS).

## Prerequisites

- **Node.js** >= 20.19
- **Docker** + Docker Compose (for infrastructure)
- **Go** (for backend and tools)
- **Nx** (installed via `npm install`)

## Quick Start

```bash
# 1. Start infrastructure (PostgreSQL, Redpanda, RabbitMQ, Jaeger)
docker compose up postgres redpanda rabbitmq jaeger

# 2. Create the database table (run once)
psql postgres://postgres:password@localhost:55000/flatty -f db.sql

# 3. Install dependencies
npm install

# 4. Run everything
npm run dev
```

## NPM Commands

| Command | Description |
|---------|-------------|
| `npm run dev` | Run all apps in parallel |
| `npm run dev:frontend` | Run launcher + settings + resident + wrapper |
| `npm run dev:backend` | Run go-api + nest-graphql |
| `npm run dev:resident` | Run launcher + resident |
| `npm run dev:launcher` | Run launcher only |
| `npm run build` | Build all apps |
| `npm run test` | Test all apps |
| `npm run lint` | Lint all apps |
| `npm run graph` | Visualize Nx dependency graph |

## Code Generation

### REST types (Swagger)

```bash
npm run api -w @flatty-budget/sdk
```

Generates `packages/sdk/src/types/api.ts` from the running Go API's Swagger spec.

### GraphQL types

```bash
npm run graphql:codegen -w @flatty-budget/sdk
```

Generates `packages/sdk/src/types/graphql.ts` from the NestJS GraphQL schema.

## Architecture

```
┌─────────────────────────────────────────────────────────────────────┐
│                        Shell Hosts                                  │
│  ┌───────────────────┐          ┌──────────────────────────────┐    │
│  │  react-launcher   │          │  react-wrapper               │    │
│  │  (port 9000)      │          │  (port 5174)                 │    │
│  │  Single MFE demo  │          │  Full shell with nav         │    │
│  └───────┬───────────┘          └────────┬─────────────────────┘    │
│          │   Script injection            │ Script injection         │
│          ▼                               ▼                          │
│  ┌──────────────────────────────────────────────────────────────┐   │
│  │              Webpack UMD Micro-Frontends                     │   │
│  │  ┌───────────────────┐     ┌────────────────────────────┐    │   │
│  │  │  react-settings   │     │  react-resident            │    │   │
│  │  │  (port 8081)      │     │  (port 8082)               │    │   │
│  │  │  User settings    │     │  Resident management       │    │   │
│  │  └───────────────────┘     └────────────────────────────┘    │   │
│  └──────────────────────────────────────────────────────────────┘   │
│              ▲ Script injection                                     │
│              │                                                      │
│  ┌───────────┴─────────────────────────────────────────────────────┐│
│  │  react-entrypoint (Next.js, ports 3000 / 3001)                  ││
│  │  Main shell — expenses, budgets, PDF upload & parsing           ││
│  └──────────────────────────────────┬──────────────────────────────┘│
│                                     │ rewrites: /api → go-api,       │
│                                     │ /graphql → nest-graphql        │
│                                     ▼                                │
├─────────────────────────────────────────────────────────────────────┤
│                          Backend APIs                               │
│  ┌───────────────────┐          ┌────────────────────────────┐      │
│  │  go-api           │          │  nest-graphql              │      │
│  │  Go + Gin         │          │  NestJS + TypeORM          │      │
│  │  REST API         │          │  GraphQL API               │      │
│  │  port 8080        │          │  port 3000                 │      │
│  └────────┬──────────┘          └─────────┬──────────────────┘      │
│           │                               │                         │
│           └──────────┬────────────────────┘                         │
│                      ▼                    │                         │
│           ┌──────────────────────┐        │                         │
│           │  PostgreSQL          │        │                         │
│           │  (port 55000)        │        │                         │
│           └──────────────────────┘        │  OTLP traces            │
│                                           │  (http/protobuf)        │
│                                           ▼                         │
├─────────────────────────────────────────────────────────────────────┤
│                      Infrastructure (Docker)                        │
│  ┌──────────┐ ┌──────────┐ ┌──────────┐ ┌──────────────────────┐    │
│  │ Postgres │ │ Redpanda │ │ RabbitMQ │ │ Jaeger (traces)      │    │
│  └──────────┘ └──────────┘ └──────────┘ │ UI 16686 · OTLP 4318 │    │
│                                         └──────────────────────┘    │
└─────────────────────────────────────────────────────────────────────┘
```

## Project Structure

| App | Port | Tool | Role |
|-----|:----:|:----:|------|
| `apps/react-launcher` | 9000 | Vite | Shell — loads a single remote MF via script injection |
| `apps/react-wrapper` | 5174 | Vite | Shell — full app with nav, integrates remotes as pages |
| `apps/react-settings` | 8081 | Webpack | Remote — user settings page |
| `apps/react-resident` | 8082 | Webpack | Remote — resident management CRUD |
| `apps/react-entrypoint` | 3000 (3001) | Next.js | Main shell — expenses, budgets, PDF upload & parsing |
| `apps/go-api` | 8080 | Go + Gin | REST API (PostgreSQL, pgx) |
| `apps/nest-graphql` | 3000 | NestJS + TypeORM | GraphQL API (PostgreSQL) |
| `apps/nginx-proxy` | 80 | Nginx | Reverse proxy for containerized deployments |
| `packages/sdk` | — | TypeScript | Shared API client + generated types (REST + GraphQL) |

Dev servers run on the host via Nx. Docker is used only for infrastructure (PostgreSQL, Redpanda, RabbitMQ, Jaeger).

## Micro-Frontends

Remote MF apps (`react-settings`, `react-resident`) are built as **Webpack UMD** libraries,
exposed on `window.ext-apps[name]`. Shell apps (`react-launcher`, `react-wrapper`) load them
at runtime via dynamic `<script>` injection.

## PDF Parsing (Invoices)

`react-entrypoint` can parse PDF invoices and auto-fill the expense form.

**Flow:**

1. User selects a PDF on the expense create page (`/expenses/create`)
2. The file is uploaded to `POST /api/uploads` — a local Next.js Route Handler (auth-protected, 10 MB limit)
3. The PDF is saved to `apps/react-entrypoint/pdfs/` (gitignored)
4. The `invoice-parser` binary (`apps/react-entrypoint/mcps/invoice-parser`) extracts vendor, amount, date, description and category
5. Parsed data auto-fills the form (amount, month/year, description, category) — the user reviews and submits

The parser runs as a local CLI (`./invoice-parser -file <path>`). See `apps/react-entrypoint/mcps/README.md` for details.

## Observability (Jaeger)

Distributed tracing via **OpenTelemetry**. `nest-graphql` exports traces to Jaeger.

```bash
docker compose up jaeger
```

- **Jaeger UI:** http://localhost:16686
- **OTLP endpoints:** `4317` (gRPC), `4318` (HTTP)
- **Config:** `OTEL_SERVICE_NAME`, `OTEL_EXPORTER_OTLP_PROTOCOL`, `OTEL_EXPORTER_OTLP_ENDPOINT` in `docker-compose.yml` / `.env.example`

## Tools

### review

Code review CLI that analyses staged git changes via Ollama.

```bash
cd tools/review
go build -o review .

./review                          # review staged changes in shop-graphql-nestjs/
./review --dir apps/react-launcher # review staged changes in a specific directory
./review --lang go                 # review Go code (default: TypeScript)
./review --model qwen2.5-coder:14b
./review --ollama-url http://192.168.1.85:11434
./review --help
```

### sql

Finds duplicate SQL between Go and NestJS repositories.

```bash
cd tools/sql
go build -o sql .

./sql
./sql --model deepseek-coder:6.7b
./sql --ollama-url http://192.168.1.85:11434
./sql --help
```