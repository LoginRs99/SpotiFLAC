# Docker Web Build

This branch adds a Docker-hosted web version of SpotiFLAC while keeping the Wails desktop app intact.

## Run

```sh
docker compose up --build
```

Open `http://localhost:8080`.

## Test Published Image On A Server

The GitHub workflow `.github/workflows/docker-web.yml` publishes the image to GitHub Container Registry when `docker-web` is pushed, and it can also be run manually from the Actions tab.

```sh
docker pull ghcr.io/loginrs99/spotiflac-web:docker-web
docker run --rm -p 8080:8080 --tmpfs /tmp/spotiflac-web:size=4g ghcr.io/loginrs99/spotiflac-web:docker-web
```

Then open `http://SERVER_IP:8080`.

## Portainer Stack

Use this stack when deploying the published GHCR image:

```yaml
services:
  spotiflac-web:
    image: ghcr.io/loginrs99/spotiflac-web:docker-web
    container_name: spotiflac-web
    ports:
      - "8080:8080"
    environment:
      SPOTIFLAC_HOST: "0.0.0.0"
      SPOTIFLAC_PORT: "8080"
      SPOTIFLAC_TMP_DIR: "/tmp/spotiflac-web"
    tmpfs:
      - /tmp/spotiflac-web:size=4g
    healthcheck:
      test: ["CMD", "spotiflac-web", "--healthcheck"]
      interval: 30s
      timeout: 5s
      start_period: 20s
      retries: 3
    restart: unless-stopped
```

When updating, redeploy the stack with image pulling enabled so Portainer fetches the newest `docker-web` tag.

## Health And Status

Useful endpoints:

- `GET /api/health`: lightweight healthcheck with version, uptime, FFmpeg status, queue counts, and temp directory.
- `GET /api/docker/status`: fuller Docker status payload with app/runtime/queue details.

The image also includes a Docker `HEALTHCHECK`, so Portainer should show healthy/unhealthy state after startup.

## Download Behavior

The Docker web build does not keep a permanent music library in the container. Each download runs in a temporary job directory, then the completed file is streamed to the browser with an attachment response so the user saves it on their own PC.

The temporary job directory is removed after the response finishes. In the provided compose file it is mounted as tmpfs at `/tmp/spotiflac-web`.

On startup, the server removes stale `download-*` directories from the temp directory in case the previous container stopped during an active download.

## Configuration

Environment variables:

- `SPOTIFLAC_HOST`: bind host inside the container, default `0.0.0.0`.
- `SPOTIFLAC_PORT`: web server port inside the container, default `8080`.
- `SPOTIFLAC_TMP_DIR`: temporary download workspace, default `/tmp/spotiflac-web`.

The compose file publishes host port `${SPOTIFLAC_PORT:-8080}` to container port `8080`.

## Upstream Sync Flow

Keep `main` close to upstream and put Docker-specific work on `docker-web`.

```sh
git switch main
git fetch upstream
git merge --ff-only upstream/main
git push origin main

git switch docker-web
git rebase main
git push origin docker-web --force-with-lease
```

See [upstream-sync.md](upstream-sync.md) for the full sync workflow and conflict guidance.

## Known V1 Limits

The Docker web build focuses on core search, metadata, settings, queue, and download workflows. Desktop-only features such as native folder pickers, opening folders in the OS file manager, drag-and-drop local file tools, audio conversion, resampling, and file renaming are stubbed or return empty results in the browser build.
