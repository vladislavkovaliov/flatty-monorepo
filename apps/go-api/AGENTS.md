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
| `nx test go-api` | Run tests (`go test ./...`) |
| `nx lint go-api` | Lint (`staticcheck ./...`) |
| `nx hot go-api` | Hot reload (`air`) |
