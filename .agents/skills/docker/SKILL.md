---
name: docker
description: Docker patterns for Cubby Go + SQLite
---

## When to use

Use for building and running the Cubby Go backend in containers, including SQLite storage and ARM64 builds.

## Base image decision

Default: `gcr.io/distroless/static-debian12:nonroot` for production.
Alternative: `alpine` only when a shell or debugging tools are required.

## Production multi-stage Dockerfile

```dockerfile
# syntax=docker/dockerfile:1.7
FROM golang:1.25-alpine3.20 AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -ldflags="-s -w" -o /app/cubby ./cmd/server

FROM gcr.io/distroless/static-debian12:nonroot
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /usr/share/zoneinfo /usr/share/zoneinfo
COPY --from=builder /app/cubby /app/cubby

USER nonroot:nonroot
VOLUME ["/data"]
EXPOSE 8080
HEALTHCHECK --interval=30s --timeout=3s --start-period=10s --retries=3 \
  CMD ["/app/cubby", "-healthcheck"]
ENTRYPOINT ["/app/cubby"]
```

## Layer ordering for cache

Copy `go.mod`/`go.sum` first, `go mod download`, then copy source.
Use BuildKit cache mounts for `/go/pkg/mod` and `/root/.cache/go-build`.

## Go build flags

- `CGO_ENABLED=0` for modernc.org/sqlite
- `-ldflags="-s -w"` to reduce binary size
- `GOOS=linux GOARCH=amd64|arm64` explicit

## Cross-compilation (ARM64) with buildx

```bash
docker buildx build \
  --platform linux/amd64,linux/arm64 \
  -t registry/cubby:1.0.0 \
  --push \
  ./backend
```

## SQLite volume handling (WAL)

SQLite WAL uses `db`, `db-wal`, `db-shm` on the same volume.
Mount a single `/data` volume and point `DB_PATH` to `/data/cubby.db`.
Ensure volume is writable by UID 65534 (distroless nonroot).

## Security hardening checklist

- Run as nonroot (`USER nonroot:nonroot`)
- No shell in runtime image
- `ENTRYPOINT` in exec form
- Drop Linux caps where supported (`cap_drop: ["ALL"]`)

## Docker Compose for development

```yaml
version: "3.8"
services:
  cubby:
    build:
      context: ./backend
    ports:
      - "8080:8080"
    volumes:
      - ./backend:/app
      - cubby_data:/data
    environment:
      DB_PATH: /data/cubby.db
      DB_WAL_MODE: "true"
    command: go run ./cmd/server
    healthcheck:
      test: ["CMD", "/app/cubby", "-healthcheck"]
      interval: 10s
      timeout: 3s
      start_period: 5s
      retries: 3
volumes:
  cubby_data:
```

## HEALTHCHECK pattern

Use exec form and a dedicated flag or endpoint.
Avoid shell form on distroless images.

## .dockerignore essentials

```text
.git
.env
bin/
frontend/node_modules/
frontend/dist/
*_test.go
*.db
*.db-wal
*.db-shm
```

## Image size comparison

| Base | Size | Notes |
| --- | --- | --- |
| scratch | smallest | no certs, no tzdata |
| distroless | small | certs + tzdata, no shell |
| alpine | larger | shell available |

## Rules

- Use distroless for production unless debugging requires alpine.
- Keep WAL files on the same mounted volume.
- Always set `CGO_ENABLED=0` for SQLite.
- Pin base image versions; avoid `latest`.

## Don'ts

- Don’t split DB/WAL/SHM across different volumes.
- Don’t run the runtime image as root.
- Don’t use shell-form ENTRYPOINT/CMD on distroless.
