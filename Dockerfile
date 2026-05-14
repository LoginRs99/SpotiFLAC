# syntax=docker/dockerfile:1

FROM node:24-bookworm AS frontend
WORKDIR /src/frontend
RUN corepack enable && corepack prepare pnpm@9 --activate
COPY frontend/package.json frontend/pnpm-lock.yaml ./
RUN pnpm install --frozen-lockfile --ignore-scripts
COPY frontend ./
COPY wails.json ../wails.json
RUN pnpm run generate-icon && pnpm run build

FROM golang:1.26-bookworm AS backend
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
COPY --from=frontend /src/frontend/dist ./frontend/dist
RUN CGO_ENABLED=1 go build -tags dockerweb -o /out/spotiflac-web .

FROM debian:bookworm-slim AS runtime
WORKDIR /app
RUN apt-get update \
    && apt-get install -y --no-install-recommends ca-certificates ffmpeg \
    && rm -rf /var/lib/apt/lists/*
COPY --from=backend /out/spotiflac-web /usr/local/bin/spotiflac-web
COPY --from=frontend /src/frontend/dist ./frontend/dist
ENV SPOTIFLAC_HOST=0.0.0.0
ENV SPOTIFLAC_PORT=8080
ENV SPOTIFLAC_TMP_DIR=/tmp/spotiflac-web
EXPOSE 8080
CMD ["spotiflac-web"]
