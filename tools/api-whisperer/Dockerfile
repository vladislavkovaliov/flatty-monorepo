FROM golang:1.26-alpine AS builder
WORKDIR /build
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=darwin GOARCH=arm64 go build -o /api-whisperer ./cmd/api-whisperer

FROM alpine:3.21
RUN apk add --no-cache ca-certificates
COPY --from=builder /api-whisperer /api-whisperer
