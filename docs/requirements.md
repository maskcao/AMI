# AMI Business Requirements (Summary)

## Goal

Build an attack knowledge intelligence platform that:

1. Ingests global malware, APT, and ICS attack reports
2. Extracts and maps attacker TTPs to MITRE ATT&CK (Enterprise or ICS by target industry)
3. Stores structured knowledge for future AI-assisted analysis

Design detail: **[design.md](./design.md)** (lightweight; **attack-knowledge-centric**, not IOC-centric like OpenCTI).  
Data model: **[data-model.md](./data-model.md)**.

**Product center:** `AttackKnowledge`（攻击知识）. **Intel reports**（情报）and **entities** hang under knowledge; IOC is subsidiary evidence.

## Out of scope (current phase)

- Enterprise control measures / defense mapping
- Multi-tenant
- Embedded SecClaw deployment (optional API only, later)
- Full OpenCTI-style graph DB and connector marketplace

## MVP (Phase 1)

- Intel list and detail UI (see [ui-spec.md](./ui-spec.md))
- Target industry dimension drives ATT&CK matrix selection
- Manual enrichment workflow (demo uses mock data)
- `webui/` business demo; `service/` placeholder

## Phase 2 (planned)

- PostgreSQL + REST API
- Entity library pages (actors, malware) with report pivot
- Ingest jobs, review queue, IOC tables
- Optional STIX import/export subset

## Repository layout

- `docs/` — specifications and design
- `webui/` — Vue 3 frontend demo
- `service/` — backend (future)
