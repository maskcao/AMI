# AMI Data Model (Attack-Knowledge-Centric)

Field-level model for `service/`. See [design.md](./design.md) §5 for philosophy vs OpenCTI.

**Core:** `AttackKnowledge`  
**Evidence:** `IntelReport` + `intel_artifact` / `intel_parsed`  
**Reference:** entities; **subsidiary:** IOC on reports

**Read rule:** UI list/query uses **SQLite columns only**; raw files via `storage_path` on demand. See [storage-and-artifacts.md](./storage-and-artifacts.md).

---

## AttackKnowledge（攻击知识 — 核心）

| Field | Type | Notes |
|-------|------|-------|
| `id` | int | PK |
| `title` | string | 知识标题 |
| `summary` | text | 沉淀概述（DB，列表/详情默认读此） |
| `attack_matrix` | enum | Enterprise \| ICS |
| `target_industries` | string[] | |
| `techniques[]` | | ATT&CK 攻击手法 |
| `tools[]` | | 攻击工具 { name, category, description } |
| `infrastructure[]` | | 攻击者基础设施 { type, value, role, confidence } |
| `vulnerabilities[]` | | CVE { cveId, description, inWild } |
| `malware[]` | | { name, class } — class 含 **ransomware** 等 |
| `threat_actors[]` | | |
| `quality_score` | int | 含各维度完整度 |
| `status` | enum | draft / published / archived |
| `needs_review` | bool | 待复核（列表筛选） |
| `updated_at` | datetime | 列表排序 |
| `linked_intel_count` | int | 可选反范式，挂载/解绑时维护 |

Relations: intel via `knowledge_intel_link` (`supports`).

**Indexes (list):** `(status, updated_at DESC)`, `(needs_review, updated_at DESC)`.

Malware classes: see `webui/src/constants/malware.ts`.

---

## IntelReport（情报 — 证据层）

Demo type: `IntelEntry` / `IntelDetail` in `webui/src/types/intel.ts`.

| Field | Type | Required | Notes |
|-------|------|----------|-------|
| `id` | int | PK | 贯穿 Raw/Parsed/链接 |
| `title` | string | yes | 报告原标题；合并规则见 design §6.5 |
| `source_id` | int | FK | → Source |
| `source_url` | string | no | |
| `source_name` | string | no | 显示名（原 `collect_source_name`） |
| `disclosed_at` | datetime | no | |
| `collected_at` | datetime | yes | |
| `updated_at` | datetime | yes | **列表排序** |
| `summary_snippet` | string | no | Parse 后写入，**列表摘要** |
| `body_preview` | text | no | 可选 ≤2KB，详情快速预览 |
| `original_body` | text | no | **废弃为主存储**；可保留兼容 Mock |
| `ingest_status` | enum | queued / parsing / extracting / succeeded / failed / archived |
| `knowledge_action` | enum | none / created / updated |
| `extract_confidence` | float | no | AI 0–1 |
| `pipeline_error` | string | no | |
| `parse_version` | string | no | 当前 Parsed 版本，如 `v1` |
| `has_raw` | bool | no | 是否有物证文件 |
| `has_parsed` | bool | no | 是否已完成清洗 |
| `archived_at` | datetime | no | 非空则默认列表排除 |
| `knowledge_ids` | int[] | | 逻辑字段；持久化用 link 表 |

**Indexes (list):** `(updated_at DESC)`, `(ingest_status, updated_at DESC)`, `(source_id, disclosed_at)`.

**Not on IntelReport:** primary TTP bindings → `AttackKnowledge`.

---

## intel_artifact（原始物证索引）

| Field | Type | Notes |
|-------|------|-------|
| `intel_id` | int | FK IntelReport |
| `seq` | int | 1,2,3… 对应 `raw/001.*` |
| `role` | enum | html / pdf / image / attachment |
| `storage_path` | string | 相对 `AMI_DATA_DIR` |
| `sha256` | string | |
| `size_bytes` | int | |
| `mime` | string | no |
| `fetched_at` | datetime | |

**PK:** `(intel_id, seq)`. **API:** 列表不返回；`GET /intel/:id/artifacts` 按需。

---

## intel_parsed（清洗结果索引）

| Field | Type | Notes |
|-------|------|-------|
| `intel_id` | int | FK |
| `version` | string | e.g. `v1` |
| `body_md_path` | string | |
| `body_txt_path` | string | |
| `chunk_path` | string | `chunks.jsonl` |
| `parser_profile_id` | uuid | |
| `parsed_at` | datetime | |

**PK:** `(intel_id, version)`.

---

## knowledge_intel_link

| Field | Type | Notes |
|-------|------|-------|
| `knowledge_id` | int | FK AttackKnowledge |
| `intel_id` | int | FK IntelReport |
| `role` | enum | primary / supporting |
| `linked_at` | datetime | |

**Indexes:** `(knowledge_id)`, `(intel_id)`.

---

## ingest_run / CollectJob

| Field | Type | Notes |
|-------|------|-------|
| `id` | uuid | |
| `collect_source_id` | uuid | |
| `started_at` / `finished_at` | datetime | |
| `status` | enum | |
| `stats_json` | json | 成功/失败条数 |

见 [system-config.md](./system-config.md). 长期运行建议按月归档。

---

## llm_response_cache

| Field | Type | Notes |
|-------|------|-------|
| `prompt_hash` | string | PK |
| `model` | string | |
| `response_json` | text | |
| `expires_at` | datetime | |

---

## llm_call_log

| Field | Type | Notes |
|-------|------|-------|
| `id` | int | PK |
| `intel_id` | int | nullable |
| `stage` | string | parse / enrich / classify |
| `model` | string | |
| `prompt_hash` | string | |
| `tokens_in` / `tokens_out` | int | |
| `created_at` | datetime | |

---

## knowledge_corpus_export

| Field | Type | Notes |
|-------|------|-------|
| `knowledge_id` | int | FK |
| `path` | string | JSONL 相对路径 |
| `format` | string | `jsonl` |
| `created_at` | datetime | |

---

## analytics_rollup（可选，SQLite）

预计算维度统计，供威胁情报库 `/analytics/*`，避免运行时大聚合。

| Field | Type | Notes |
|-------|------|-------|
| `dimension` | string | cve / malware / … |
| `bucket_key` | string | |
| `count` | int | |
| `refreshed_at` | datetime | |

或由 DuckDB 承担，见 [storage-and-artifacts.md](./storage-and-artifacts.md) §10.

---

## Full-text search (optional)

Virtual tables: `intel_report_fts`, `attack_knowledge_fts` — 同步 `title`, `summary`, `summary_snippet`. 关键词检索走 FTS，避免 `%like%` 扫主表。

---

## ThreatActor / Malware / MitreTechnique / Procedure / Ioc

- **Procedure**, **KnowledgeTechnique**: FK → `attack_knowledge_id`
- **Ioc**: FK → `intel_report_id`; index `(value)`, `(intel_report_id)`
- **ThreatActor / Malware**: M:N with `AttackKnowledge` (`involves`)

---

## pipeline_jobs

| Field | Type | Notes |
|-------|------|-------|
| `id` | int | PK |
| `intel_id` | int | |
| `stage` | enum | ingest / parse / extract |
| `status` | enum | pending / running / done / failed |
| `run_at` | datetime | |
| `locked_at` | datetime | Worker 抢锁 |

**Index:** `(status, run_at)`.

---

## STIX Export (Phase 2)

- Export **AttackKnowledge** as custom extension or aggregated `grouping`
- Export **IntelReport** as STIX `report` with `object_refs`
- Do **not** treat AMI as indicator-first STIX feed
