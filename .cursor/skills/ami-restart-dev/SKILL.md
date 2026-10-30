---
name: ami-restart-dev
description: Restarts AMI Go API (port 8080) and Vite WebUI (port 5173), stops stale processes, verifies health, and restores missing source from git when needed. Use when the user asks to restart, start, or debug frontend/backend services, mentions 重启/启动/调试服务, or needs dev servers running for this repository.
---

# AMI 开发服务重启

Monorepo layout: `service/` (Go API), `webui/` (Vue + Vite). Run backend and frontend in **two separate background terminals**.

## Quick restart

### 1. Stop existing processes

```bash
pkill -f 'go run ./cmd/ami' 2>/dev/null || true
pkill -f '/ami/service.*cmd/ami' 2>/dev/null || true
pkill -f 'vite --host' 2>/dev/null || true
sleep 1
ss -tlnp | grep -E ':(5173|8080)' || echo 'ports 5173/8080 free'
```

### 2. Preflight (only if start fails)

| Check | Command / action |
|-------|------------------|
| Go entry exists | `test -f service/cmd/ami/main.go` |
| WebUI manifest exists | `test -f webui/package.json` |
| Restore deleted tracked source | See [Restore source](#restore-source) |
| WebUI env | `test -f webui/.env.development \|\| cp webui/.env.development.example webui/.env.development` |
| Node deps | `test -d webui/node_modules \|\| (cd webui && npm install)` |

### 3. Start backend (terminal 1, background)

From repo root:

```bash
cd service
AMI_ADDR=0.0.0.0:8080 AMI_DATA_DIR=./data AMI_SEED=1 go run ./cmd/ami -seed
```

Equivalent via Makefile (same env):

```bash
cd service && make run
```

Success log: `AMI API listening on 0.0.0.0:8080 (data: ./data)`.

### 4. Start frontend (terminal 2, background)

```bash
cd webui
npm run dev
```

Success log: `VITE ... ready` on `http://localhost:5173/` (also `0.0.0.0:5173`).

### 5. Verify

```bash
curl -s http://127.0.0.1:8080/health          # {"status":"ok"}
curl -s -o /dev/null -w '%{http_code}\n' \
  'http://127.0.0.1:8080/api/v1/intel?page=1&page_size=1'   # 200
curl -s -o /dev/null -w '%{http_code}\n' http://127.0.0.1:5173/   # 200
```

API base path is **`/api/v1`** (not `/api` alone). WebUI proxies `/api` → `http://127.0.0.1:8080` (see `webui/vite.config.ts`).

## Environment variables

| Variable | Dev default | Purpose |
|----------|-------------|---------|
| `AMI_ADDR` | `0.0.0.0:8080` | API listen address |
| `AMI_DATA_DIR` | `./data` (under `service/`) | SQLite + artifact store |
| `AMI_SEED` | `1` with `make run` | Enables `-seed` path when combined with flag |
| `-seed` flag | passed in `make run` | Demo seed **only if DB is empty** (`seed.Run`) |

Optional tuning: `AMI_WORKER_INTERVAL_SEC`, `AMI_INGEST_TICK_SEC`, `AMI_INGEST_MIN_GAP_SEC` (`service/internal/config/config.go`).

## WebUI API mode

`webui/.env.development` should exist with:

- `VITE_USE_API=true`
- `VITE_API_BASE=` (empty; use Vite proxy)

Copy from example if missing: `cp webui/.env.development.example webui/.env.development`.

## Restore source

If `find service -name '*.go' | wc -l` is 0 or `webui/package.json` is missing, tracked files were likely deleted locally. Restore **code only** (keep `service/data`):

```bash
git checkout HEAD -- \
  service/cmd service/Makefile service/go.mod service/go.sum service/internal \
  webui/package.json webui/package-lock.json webui/vite.config.ts webui/index.html \
  webui/tsconfig.json webui/tsconfig.node.json webui/env.d.ts \
  webui/.env.development webui/.env.development.example webui/.gitignore \
  webui/public webui/src
```

Then repeat [Quick restart](#quick-restart).

## LAN access

- WebUI: `http://<host-ip>:5173/` (`vite --host 0.0.0.0`, port **5173**)
- API direct: `http://<host-ip>:8080` (port **8080**)
- Browser API calls: `http://<host-ip>:5173/api/v1/...` (proxied to backend)

Common routes: see [reference.md](reference.md).

## Reporting to the user

After restart, state:

1. Backend URL and health result
2. Frontend URL (localhost + LAN IP if available via `hostname -I`)
3. Any preflight fix applied (restore source, npm install, env copy)

Do **not** run `git checkout` over `service/data` unless the user explicitly asks to reset data.
