# go-api

REST API built with Go + Gin.

- **Port:** 8080
- **Stack:** Go, Gin
- **Entry:** `cmd/api/main.go`

## Commands

| Command | Description |
|---------|-------------|
| `nx serve go-api` | Start server |
| `nx build go-api` | Build (`go build -o dist/apps/go-api/api ./cmd/api/`) |
| `nx test go-api` | Run all tests (`go test ./...`) |
| `nx lint go-api` | Lint (`staticcheck ./...`) |
| `nx hot go-api` | Hot reload (`air`) |
| `nx migrate-db go-api` | Apply DB migrations (goose) |

## Database

Schema has exactly one owner: **goose**. `db.sql` is the base schema (tables + seed
data), loaded once by the compose `init-db` service. Everything on top of it lives in
`migrations/*.sql`, embedded into the binary via `//go:embed` (`migrations/embed.go`)
and applied by `cmd/migrate`.

```bash
nx migrate-db go-api              # up to latest
go run ./cmd/migrate -status      # list, with version
go run ./cmd/migrate -down        # roll back one
```

Re-running is safe: goose records applied versions in `goose_db_version`. Migrations
`006`/`007` reference the better-auth `"user"` table, so they need `nest-graphql`
running (or the compose `migrate` service, which waits for its healthcheck).

## Testing

Unit tests live alongside the code as `*_test.go` files. They use the standard `testing` package with `testify/mock` for mocks and `testify/assert` for assertions — no external infrastructure required.

```bash
# Run all tests (unit + integration — integration excluded by build tags by default)
nx test go-api

# Run tests directly
go test ./...

# Run a specific package
go test ./repos/user/...

# Run a specific test
go test -run TestPgxRepository_List ./repos/user/...

# Run tests verbosely (shows subtest names)
go test -v ./repos/user/...

# Run with race detection
go test -race ./...
```

> **Unit tests** use mocked database connections and run in milliseconds with zero infrastructure.
> **Integration tests** (not yet implemented) should use `//go:build integration` build tags and `-tags=integration` to run.
