# API Overview (Draft)

Base: `/api/v1`. **Attack-knowledge-first** (unlike OpenCTI indicator-first workflows).

## Attack knowledge（核心）

| Method | Path | Description |
|--------|------|-------------|
| GET | `/knowledge` | List knowledge (filter: industry, TTP, actor, malware, q) |
| GET | `/knowledge/:id` | Detail + techniques, entities, linked intel reports |
| POST | `/knowledge` | Create knowledge |
| PATCH | `/knowledge/:id` | Update |
| POST | `/knowledge/:id/publish` | Publish |
| POST | `/knowledge/:id/intel` | Link intel report(s) (`supports`) |
| DELETE | `/knowledge/:id/intel/:intelId` | Unlink |

## Intel reports（情报 / 证据）

| Method | Path | Description |
|--------|------|-------------|
| GET | `/intel` | List reports (ingest workflow) |
| GET | `/intel/:id` | Detail + original body, IOCs, linked knowledge ids |
| POST | `/intel/import` | Upload / URL ingest |
| POST | `/intel/:id/publish` | Publish report |

Phase 1 demo implements `/intel` only; shapes in `webui/src/types/intel.ts`.

## Entities

| Method | Path | Description |
|--------|------|-------------|
| GET | `/actors/:id/knowledge` | Knowledge involving actor |
| GET | `/malware/:id/knowledge` | Knowledge involving malware |

## IOC（辅助）

| Method | Path | Description |
|--------|------|-------------|
| GET | `/ioc` | Search → intel reports + related knowledge (not primary UX) |

## Search

| Method | Path | Description |
|--------|------|-------------|
| GET | `/search?type=knowledge\|intel\|ioc` | Default `type=knowledge` |

## Types

See [data-model.md](./data-model.md).
