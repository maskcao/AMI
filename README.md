# AMI — 攻击知识情报平台

Attack knowledge intelligence platform: ingest threat reports, extract TTPs (MITRE ATT&CK), and build a structured knowledge base for analysis and future AI workflows.

## Repository layout

| Directory | Description |
|-----------|-------------|
| [`docs/`](./docs/) | Requirements, UI spec, industry taxonomy, API draft |
| [`webui/`](./webui/) | Vue 3 business demo (mock data) |
| [`service/`](./service/) | Backend placeholder (not implemented) |

## Quick start (webui demo)

```bash
cd webui
npm install
npm run dev
```

Open http://localhost:5173 — default home **attack knowledge** at `/knowledge`.

## Demo features

- **Attack knowledge** (`/knowledge`): core list — TTP, industries, matrix, linked intel count
- **Knowledge detail**: analysis tab + **支撑情报** tab (evidence reports)
- **Intel workbench** (`/intel`): ingest status, IOC count, links to knowledge IDs
- **Intel detail**: linked knowledge cards + IOC (subsidiary) + original text
- **Dashboard** (`/dashboard`): first-level analytics home
- **Analytics dimensions** (`/analytics/cve`, …): CVE, malware, actors, MITRE, tools, infra, IOC
- **IOC search** (`/ioc`): auxiliary lookup → intel → knowledge (not the home pivot)
- Knowledge **1001** has **2** intel reports (multi-source demo)
- Mock data only; no `service` API yet

## Documentation

See [docs/README.md](./docs/README.md). Platform design (OpenCTI-inspired, lightweight): [docs/design.md](./docs/design.md).
