FROM golang:1.26-alpine AS builder
WORKDIR /build
COPY apps/go-api/go.mod apps/go-api/go.sum ./
RUN go mod download
COPY apps/go-api/ .
RUN CGO_ENABLED=0 go test ./...
RUN CGO_ENABLED=0 go build -o api ./cmd/api/
RUN CGO_ENABLED=0 go build -o migrate ./cmd/migrate/

FROM alpine:3.21
RUN apk add --no-cache ca-certificates curl
COPY --from=builder /build/api /api
COPY --from=builder /build/migrate /migrate
EXPOSE 8080
HEALTHCHECK --interval=30s --timeout=5s --start-period=10s --retries=3 \
  CMD curl -f http://localhost:8080/api/health || exit 1
CMD ["/api"]
