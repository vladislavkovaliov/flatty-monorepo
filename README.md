# flatty-budget-monorepo

Nx monorepo with micro-frontends (React) + backends (Go/NestJS).

## Prerequisites

- **Node.js** >= 20.19
- **Docker** + Docker Compose — or **podman** + podman-compose (for infrastructure)
- **Go** (for backend and tools)
- **Nx** (installed via `npm install`)

## Quick Start

```bash
# 1. Start infrastructure (PostgreSQL, Redpanda, RabbitMQ, Jaeger)
docker compose --env-file .env.dev up -d postgres redpanda rabbitmq jaeger

# 2. Create the database table (run once)
psql postgres://postgres:password@localhost:55000/flatty -f db.sql

# 3. Install dependencies
npm install

# 4. Run everything
npm run dev
```

**Docker environments:**

| Mode | Command |
|------|---------|
| Dev | `docker compose --env-file .env.dev up -d` |
| Prod | `docker compose --env-file .env.prod up -d` |

## NPM Commands

| Command | Description |
|---------|-------------|
| `npm run dev` | Run all apps in parallel |
| `npm run dev:frontend` | Run launcher + settings + resident + entrypoint + vue-torrent-bot |
| `npm run dev:backend` | Run go-api + nest-graphql |
| `npm run dev:resident` | Run launcher + resident |
| `npm run dev:launcher` | Run launcher only |
| `npm run build` | Build all apps |
| `npm run test` | Test all apps |
| `npm run lint` | Lint all apps |
| `npm run graph` | Visualize Nx dependency graph |

## Backends only (podman)

Run just the APIs (**go-api**, **nest-graphql**) with their infrastructure via rootless podman.
Frontends and nginx are not started.

### One-time host setup

Rootless podman on **btrfs** needs `fuse-overlayfs`, and short image names such as
`postgres`/`rabbitmq` must be resolvable. This host's running kernel also has no `tun`
module, so pasta/slirp4netns fail — every service uses `network_mode: host` and talks over
`localhost`.

```bash
# 1. overlay storage driver on btrfs
sudo pacman -S --noconfirm fuse-overlayfs
# ~/.config/containers/storage.conf:
#   [storage]
#   driver = "overlay"
#   [storage.options.overlay]
#   mount_program = "/usr/bin/fuse-overlayfs"

# 2. resolve short image names
# ~/.config/containers/registries.conf:
#   unqualified-search-registries = ["docker.io"]
```

### Start / stop

```bash
# First run only — builds the images (go test + npm ci). Build needs host networking.
podman-compose -f docker-compose.backend.yml --podman-build-args="--network=host" build

podman-compose -f docker-compose.backend.yml up -d
podman-compose -f docker-compose.backend.yml logs -f go-api nest-graphql
podman-compose -f docker-compose.backend.yml down
```

Endpoints (host networking → plain `localhost`):

| Service | URL |
|---------|-----|
| go-api | http://localhost:8080/api/health |
| nest-graphql | http://localhost:3000/api/config |
| PostgreSQL | localhost:55000 |
| Redpanda (Kafka) | localhost:9092 |
| RabbitMQ | localhost:5672 (UI http://localhost:15672) |

### Resource limits

Each service is capped via `deploy.resources` (podman-compose maps it to
`--cpus` / `--memory` / `--memory-reservation`):

| Service | cpus | memory | reservation |
|---------|------|--------|-------------|
| postgres | 0.50 | 512M | 256M |
| redpanda | 1.00 | 1G | 256M |
| rabbitmq | 0.50 | 384M | 192M |
| go-api | 0.50 | 256M | 128M |
| nest-graphql | 1.00 | 512M | 256M |

Live usage: `podman stats`. Redpanda (seastar) crashes with less than one CPU, hence
`--smp 1` and `cpus: "1.00"`.

### python-torrent-bot (on host)

The bot's HTTP API is intentionally restricted to localhost, so run it on the host against the
podman Postgres:

```bash
cp apps/python-torrent-bot/.env.example apps/python-torrent-bot/.env   # set BOT_TOKEN
nx setup python-torrent-bot   # once: .venv + dependencies
nx serve python-torrent-bot   # API on http://127.0.0.1:8090/docs
```

`.env`: `DATABASE_URL=postgresql://postgres:password@localhost:55000/torrent_bot?sslmode=disable`,
`TEST_MODE=true`.

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
│  │          Micro-Frontends (Webpack UMD + Vite IIFE)           │   │
│  │  ┌───────────────────┐     ┌────────────────────────────┐    │   │
│  │  │  react-settings   │     │  react-resident            │    │   │
│  │  │  (port 8081)      │     │  (port 8083)               │    │   │
│  │  │  User settings    │     │  Resident management       │    │   │
│  │  └───────────────────┘     └────────────────────────────┘    │   │
│  │          vue-torrent-bot (8084, Vite IIFE)                   │   │
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
| `apps/vue-torrent-bot` | 8084 | Vite | Remote — Vue torrent-bot UI (IIFE) |
| `apps/react-entrypoint` | 3000 (3001) | Next.js | Main shell — expenses, budgets, PDF upload & parsing |
| `apps/go-api` | 8080 | Go + Gin | REST API (PostgreSQL, pgx) |
| `apps/nest-graphql` | 3000 | NestJS + TypeORM | GraphQL API (PostgreSQL) |
| `apps/nginx-proxy` | 80 | Nginx | Reverse proxy for containerized deployments |
| `packages/sdk` | — | TypeScript | Shared API client + generated types (REST + GraphQL) |

Dev servers run on the host via Nx. Docker is used only for infrastructure (PostgreSQL, Redpanda, RabbitMQ, Jaeger).

## Micro-Frontends

Remotes are built as libraries and loaded by the shells at runtime via dynamic `<script>` injection:

- **React remotes** (`react-settings`, `react-resident`) — Webpack **UMD**, exposed as `window.apps[name]`.
- **Vue remote** (`vue-torrent-bot`) — Vite **IIFE**, exposes `window["ext-apps"]` implementing
  `IAppComponent` (`initialize` / `destroy`) from `@flatty-budget/mfe-host`.

Shells (`react-launcher`, `react-entrypoint`) fetch `GET /api/applications` at boot and load each row's
`bundle_js` through its `proxy_base_path` rewrite.

```bash
nx serve vue-torrent-bot       # build + serve the bundle on :8084 (the host fetches /vue-torrent-bot.js)
nx storybook vue-torrent-bot   # component workshop (Storybook) on :6006
```

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

Distributed tracing via **OpenTelemetry**. `nest-graphql` and `go-api` export traces to Jaeger.

```bash
docker compose up jaeger
```

- **Jaeger UI:** http://localhost:16686
- **OTLP endpoints:** `4317` (gRPC), `4318` (HTTP)
- **Config:** `OTEL_SERVICE_NAME`, `OTEL_EXPORTER_OTLP_PROTOCOL`, `OTEL_EXPORTER_OTLP_ENDPOINT`, `OTEL_EXPORTER_OTLP_TRACES_ENDPOINT` in `docker-compose.yml` / `.env.example`

## Performance Testing (k6)

Automated load testing with k6, Prometheus, and Grafana. Scripts live in `tools/load-test/`.

### Prerequisites

- k6 >= 0.50 (install locally or use the Docker image)
- A running backend (Docker compose **or** `npm run dev:backend` on host)
- A load-test user — seed once:

```bash
curl -X POST http://localhost/api/auth/sign-up/email \
  -H 'Content-Type: application/json' \
  -d '{"email":"loadtest@flatty.local","password":"LoadTest123!"}'
```

### Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `K6_BASE_URL` | `http://localhost` | Backend URL to test against |
| `K6_EMAIL` | *(required)* | Load-test user email |
| `K6_PASSWORD` | *(required)* | Load-test user password |
| `K6_RESIDENT_LOCATION_ID` | `1` | Resident location ID for queries |

### Run — Docker Compose

Tests run inside the compose network (nginx → backends):

```bash
# Smoke (5 VU, 1 min)
docker compose --profile load-test run --rm k6 run /scripts/smoke.js

# Load (ramp to 50 VU) — streams results to Prometheus
docker compose --profile load-test run --rm k6 run --out experimental-prometheus-rw /scripts/load.js

# Stress / spike — on-demand
docker compose --profile load-test run --rm k6 run /scripts/stress.js
docker compose --profile load-test run --rm k6 run /scripts/spike.js
```

### Run — Local k6 (any backend URL)

Run k6 on your host and point it at any running backend (local dev, staging, etc.):

```bash
# Against dev backend (nginx on port 80)
K6_BASE_URL=http://localhost:80 \
K6_EMAIL=loadtest@flatty.local \
K6_PASSWORD=LoadTest123! \
k6 run tools/load-test/smoke.js

# Against go-api directly (port 8080, no nginx)
K6_BASE_URL=http://localhost:8080 \
K6_EMAIL=loadtest@flatty.local \
K6_PASSWORD=LoadTest123! \
k6 run tools/load-test/smoke.js

# Against a remote server
K6_BASE_URL=https://staging.example.com \
K6_EMAIL=loadtest@flatty.local \
K6_PASSWORD=LoadTest123! \
k6 run tools/load-test/load.js
```

Reports are written to `tools/load-test/reports/` (gitignored). Load-test requests
appear as traces in Jaeger beside server spans.

### Dashboards

- **Prometheus:** http://localhost:9090
- **Grafana:** http://localhost:3001 (admin / `GRAFANA_ADMIN_PASSWORD`, default `admin`) — k6 dashboard + Prometheus/Jaeger datasources
- **Jaeger:** http://localhost:16686

### Thresholds

| Scenario | p95 | p99 | error rate |
|----------|-----|-----|------------|
| smoke | < 500 ms | — | < 1% |
| load | < 800 ms | < 1500 ms | < 1% |
| stress / spike | < 2000 ms | — | < 5% |

A breached threshold exits k6 with code 1 (CI gate).

### CI

- **PRs:** smoke test against the compose stack (`.github/workflows/performance.yml`)
- **Nightly (02:00 UTC):** full load test; JSON report uploaded as an artifact; results streamed to Prometheus for trend dashboards

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