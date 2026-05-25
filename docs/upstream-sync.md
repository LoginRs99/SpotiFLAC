# Syncing With Upstream

This fork keeps upstream-friendly work on `main` and Docker-specific work on `docker-web`.

Upstream repository:

```sh
https://github.com/spotbye/SpotiFLAC.git
```

## One-Time Remote Setup

Run this once in your local clone:

```sh
git remote add upstream https://github.com/spotbye/SpotiFLAC.git
git remote -v
```

Expected remotes:

```text
origin    https://github.com/LoginRs99/SpotiFLAC.git
upstream  https://github.com/spotbye/SpotiFLAC.git
```

## Normal Sync Flow

Update `main` from upstream first:

```sh
git switch main
git fetch upstream
git merge --ff-only upstream/main
git push origin main
```

Then replay Docker changes on top:

```sh
git switch docker-web
git rebase main
git push origin docker-web --force-with-lease
```

The `docker-web` branch push rebuilds the Docker image workflow automatically.

## If Rebase Has Conflicts

Most conflicts should be in files touched by the Docker web layer:

- `Dockerfile`
- `docker-compose.yml`
- `docker_web_main.go`
- `frontend/wailsjs/...`
- Docker-only guards in `frontend/src/App.tsx`, `frontend/src/components/Sidebar.tsx`, and `frontend/vite.config.ts`

Resolve conflicts by keeping upstream desktop behavior intact and preserving Docker-only code behind `__DOCKER_WEB__` or the Go `dockerweb` build tag.

After resolving:

```sh
git add <resolved-files>
git rebase --continue
git push origin docker-web --force-with-lease
```

## Abort A Bad Rebase

If the rebase becomes messy:

```sh
git rebase --abort
```

Then inspect the upstream changes before trying again.

## Docker Image After Sync

Once the workflow succeeds, redeploy from:

```sh
ghcr.io/loginrs99/spotiflac-web:docker-web
```

In Portainer, redeploy the stack with image pull enabled so the latest image is used.

## Quick Smoke Test After Sync

After the Docker workflow passes:

```sh
curl http://SERVER_IP:8080/api/health
curl http://SERVER_IP:8080/api/docker/status
```

Then open the UI and test:

- Search or paste one Spotify track URL.
- Fetch metadata.
- Download one track and confirm the browser receives the file.
- Check Portainer logs for `[metadata]` and `[download]` entries if anything fails.
