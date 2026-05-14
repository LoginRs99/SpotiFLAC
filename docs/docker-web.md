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

## Download Behavior

The Docker web build does not keep a permanent music library in the container. Each download runs in a temporary job directory, then the completed file is streamed to the browser with an attachment response so the user saves it on their own PC.

The temporary job directory is removed after the response finishes. In the provided compose file it is mounted as tmpfs at `/tmp/spotiflac-web`.

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

## Known V1 Limits

The Docker web build focuses on core search, metadata, settings, queue, and download workflows. Desktop-only features such as native folder pickers, opening folders in the OS file manager, drag-and-drop local file tools, audio conversion, resampling, and file renaming are stubbed or return empty results in the browser build.
