# AMI Service

Go monolith: HTTP API + ingest scheduler + pipeline worker + optional LLM enrich.

## Stack

| Layer | Choice |
|-------|--------|
| Language | **Go** (aligned with SecClaw) |
| OLTP | **SQLite** (`ami.db` under `AMI_DATA_DIR`) |
| Artifacts | Local files under `AMI_DATA_DIR/store/` |
| Analytics (optional) | DuckDB single file |
| Queue | `pipeline_jobs` table (no Redis) |
| LLM | HTTP to configured `ModelProvider` |

See [../docs/design.md](../docs/design.md) §4, [../docs/storage-and-artifacts.md](../docs/storage-and-artifacts.md).

## Environment

```bash
export AMI_DATA_DIR=/var/lib/ami   # default: ./data
export AMI_LLM_API_KEY=sk-...      # OpenAI-compatible key (optional; enables llm-v1 extract)
export AMI_INGEST_TICK_SEC=300     # scheduler tick (default 300)
export AMI_INGEST_MIN_GAP_SEC=3600 # min gap between auto-runs per source (default 3600)
```

## Planned layout

```
service/
  cmd/ami/              # main: API + scheduler + worker
  internal/
    store/              # SQLite migrations, queries
    artifacts/          # manifest, raw/parsed paths
    pipeline/           # ingest, parse, enrich
    llm/                # provider client, cache, token budget
    api/                # HTTP handlers
```

## Read path

- **List / search / analytics:** SQLite only — no filesystem reads.
- **Original / artifacts:** `GET /intel/:id/original`, `/artifacts/:seq` on user demand.

## Endpoints

[../docs/api-overview.md](../docs/api-overview.md)

## Integration

- `webui` → REST when ready; demo: `webui/src/mock/*.ts`
- SecClaw: read-only export of published knowledge (no write-back)

## Quick start

```bash
cd service
make run    # AMI_DATA_DIR=./data, seed demo, :8080
make test
```

WebUI with API:

```bash
cp webui/.env.development.example webui/.env.development
# VITE_USE_API=true
cd webui && npm run dev   # proxies /api → :8080
```

## Phases

| Phase | Deliverable |
|-------|-------------|
| P1 | **Done (initial):** migrations, import, list/detail, pipeline parse/extract (rules) |
| P2 | **Done (initial):** PATCH/publish/merge knowledge, linked intel from DB, corpus JSONL on publish, intel retry |
| P2b | **Done (initial):** ingest RSS/URL, settings PATCH, manual collect run, LLM enrich when `AMI_LLM_API_KEY` set |
| P3 | LLM cache + `llm_call_log`, EnrichProfile budgets |
| P4 | Corpus JSONL on publish; optional DuckDB sync |
