# syntax=docker/dockerfile:1.7

# ── Build backend ──
FROM golang:1.25-alpine AS backend-builder
WORKDIR /build
COPY backend/go.mod backend/go.sum ./
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    go mod download
COPY backend/ .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /cubby ./cmd/server

# ── Build frontend ──
FROM node:22-alpine AS frontend-builder
WORKDIR /build
COPY frontend/package*.json ./
RUN npm ci
COPY frontend/ .
RUN npm run build

# ── Runtime (distroless nonroot) ──
FROM gcr.io/distroless/static-debian12:nonroot

WORKDIR /app
COPY --from=backend-builder /cubby /app/cubby
COPY --from=frontend-builder /build/dist /app/static

VOLUME ["/data"]
EXPOSE 8080

USER nonroot:nonroot
ENTRYPOINT ["/app/cubby"]
