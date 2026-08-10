# syntax=docker/dockerfile:1

# ---- 1. Build the frontend -------------------------------------------------
# Vite writes to ../internal/web/dist, which is where the Go build embeds it.
FROM node:22-alpine AS web

WORKDIR /app/frontend
COPY frontend/package.json frontend/package-lock.json ./
RUN npm ci

COPY frontend/ ./
RUN npm run build

# ---- 2. Build the binary ---------------------------------------------------
FROM golang:1.22-alpine AS build

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
COPY --from=web /app/internal/web/dist ./internal/web/dist

ARG VERSION=dev
# CGO off: the SQLite driver is pure Go, so the result is a static binary that
# runs on a distroless image with nothing else in it.
RUN CGO_ENABLED=0 GOOS=linux go build \
    -trimpath \
    -ldflags "-s -w -X main.version=${VERSION}" \
    -o /goportunitties ./cmd/api

# The database directory has to exist and belong to the runtime user, and there
# is no shell in the final image to create it there.
RUN mkdir -p /data

# ---- 3. Runtime ------------------------------------------------------------
FROM gcr.io/distroless/static-debian12:nonroot

COPY --from=build /goportunitties /goportunitties
COPY --from=build --chown=nonroot:nonroot /data /data

ENV APP_ENV=production \
    PORT=8080 \
    DB_PATH=/data/main.db

EXPOSE 8080
USER nonroot
VOLUME ["/data"]

ENTRYPOINT ["/goportunitties"]
