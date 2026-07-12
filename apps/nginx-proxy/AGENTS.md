# nginx-proxy

Nginx reverse proxy for Docker deployment.

- **Port:** 80 (internal)
- **Stack:** Nginx
- **Build:** Docker only

## Upstreams

| Path | Upstream | Target |
|------|----------|--------|
| `/` | angular_shell | angular-shell:4000 |
| `/api/` | go_api | go-api:8080 |
| `/graphql` | nest_graphql | nest-graphql:3000 |

## Commands

| Command | Description |
|---------|-------------|
| `docker build -t nginx-proxy .` | Build image |
