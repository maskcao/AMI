# API Overview (Draft)

Base: `/api/v1`. **Attack-knowledge-first** (unlike OpenCTI indicator-first workflows).

## Read path rules

| Traffic | SQLite | File store (`AMI_DATA_DIR/store`) |
|---------|--------|-----------------------------------|
| List / search / filter / sort / pagination | **Yes** | **No** |
| Detail (default) | **Yes** (metadata, summary, relations) | **No** |
| Original body / artifact download | Path lookup only | **Yes** (stream) |
| Analytics dimensions | SQLite rollup or DuckDB | **No** |

- List endpoints **must not** default `include=body` or read raw files.
- WebUI loads `/original` and `/artifacts` **lazily** (user opens tab).

See [design.md](./design.md) §4.1, [storage-and-artifacts.md](./storage-and-artifacts.md).

---

## Attack knowledge（核心）

| Method | Path | Description |
|--------|------|-------------|
| GET | `/knowledge` | List from DB only (filter: industry, TTP, actor, malware, q, needs_review) |
| GET | `/knowledge/:id` | Detail + techniques, entities, linked intel **summaries** (no file IO) |
| POST | `/knowledge` | Create knowledge |
| PATCH | `/knowledge/:id` | Update |
| POST | `/knowledge/:id/publish` | Publish; trigger corpus snapshot |
| POST | `/knowledge/:id/intel` | Link intel report(s) |
| DELETE | `/knowledge/:id/intel/:intelId` | Unlink |
| GET | `/knowledge/:id/corpus` | Latest training JSONL export metadata (optional) |

**Query:** keyset pagination `?cursor_updated_at=&limit=` preferred over large `offset`.

---

## Intel reports（情报 / 证据）

| Method | Path | Description |
|--------|------|-------------|
| GET | `/intel` | List from DB (`title`, `summary_snippet`, `ingest_status`, `updated_at`, …) |
| GET | `/intel/:id` | Detail metadata + IOCs + linked knowledge ids (**no** full original by default) |
| GET | `/intel/:id/original` | Stream raw or parsed body (**on demand**) |
| GET | `/intel/:id/artifacts` | Manifest from DB + paths |
| GET | `/intel/:id/artifacts/:seq` | Single artifact stream |
| POST | `/intel/import` | Upload / URL ingest → writes Raw + DB row |
| POST | `/intel/:id/retry` | Re-queue pipeline |

**Deprecated for list:** `GET /intel?include=body` — use `/original` on detail only.

Phase 1 demo: mock in `webui/src/mock/`; shapes in `webui/src/types/intel.ts`.

---

## Entities

| Method | Path | Description |
|--------|------|-------------|
| GET | `/actors/:id/knowledge` | Knowledge involving actor (DB) |
| GET | `/malware/:id/knowledge` | Knowledge involving malware (DB) |

---

## IOC（辅助）

| Method | Path | Description |
|--------|------|-------------|
| GET | `/ioc` | Search DB → intel summaries + related knowledge |

---

## Search

| Method | Path | Description |
|--------|------|-------------|
| GET | `/search?type=knowledge\|intel\|ioc` | Default `type=knowledge`; FTS-backed |

---

## Analytics（威胁情报库维度）

| Method | Path | Description |
|--------|------|-------------|
| GET | `/analytics/:dimension` | Aggregates from `analytics_rollup` or DuckDB — **no** file scan |

---

## Types

See [data-model.md](./data-model.md).
