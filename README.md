# AMI — 攻击知识情报平台

Attack knowledge intelligence platform: ingest threat reports, extract TTPs (MITRE ATT&CK), and build a structured knowledge base for analysis and future AI workflows.

## Repository layout

| Directory | Description |
|-----------|-------------|
| [`docs/`](./docs/) | Requirements, UI spec, industry taxonomy, API draft |
| [`webui/`](./webui/) | Vue 3 business demo (mock data) |
| [`service/`](./service/) | Go backend placeholder (SQLite + `AMI_DATA_DIR`, not implemented) |

## Quick start (webui demo)

```bash
cd webui
npm install
npm run dev
```

Open http://localhost:5173 — default home **attack knowledge** at `/knowledge`.

## 联调（WebUI + Go API）

```bash
# 终端 1：后端（种子数据 + 流水线）
cd service && make run

# 终端 2：前端（已配置 webui/.env.development → VITE_USE_API=true）
cd webui && npm run dev

# 验收脚本（后端启动后）
./scripts/integration-smoke.sh
```

Vite 将 `/api` 代理到 `http://127.0.0.1:8080`。若曾未启用 API 模式，修改 `.env.development` 后需**重启** `npm run dev`。

## Demo features

- **Attack knowledge** (`/knowledge`): core list — TTP, industries, matrix, linked intel count
- **Knowledge detail**: analysis tab + **支撑情报** tab (evidence reports)
- **Ingest queue** (`/intel`, 采集队列): pipeline status, IOC count, auto-linked knowledge IDs
- **Intel detail**: linked knowledge cards + IOC (subsidiary) + original text
- **Dashboard** (`/dashboard`): first-level analytics home
- **Analytics dimensions** (`/analytics/cve`, …): CVE, malware, actors, MITRE, tools, infra, IOC
- **IOC search** (`/ioc`): auxiliary lookup → intel → knowledge (not the home pivot)
- Knowledge **1001** has **2** intel reports (multi-source demo)
- Mock data only; no `service` API yet

## Data directory (when service is live)

```bash
export AMI_DATA_DIR=./data   # ami.db + store/raw|parsed|corpus
```

Lists and queries read **SQLite only**; original files load on demand in detail tabs.

## Documentation

See [docs/README.md](./docs/README.md). Architecture: [docs/design.md](./docs/design.md). Storage: [docs/storage-and-artifacts.md](./docs/storage-and-artifacts.md).
