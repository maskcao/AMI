# AMI Data Model (Attack-Knowledge-Centric)

Field-level model for `service/`. See [design.md](./design.md) §5 for philosophy vs OpenCTI.

**Core:** `AttackKnowledge`  
**Evidence:** `IntelReport` (maps to demo `IntelEntry` until Phase 2 split)  
**Reference:** entities; **subsidiary:** IOC on reports

---

## AttackKnowledge（攻击知识 — 核心）

| Field | Type | Notes |
|-------|------|-------|
| `id` | int | PK |
| `title` | string | 知识标题 |
| `summary` | text | 沉淀概述 |
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

Relations: intel via `knowledge_intel_link` (`supports`).

Malware classes: see `webui/src/constants/malware.ts` (勒索软件、后门、僵尸网络…).

---

## IntelReport（情报 — 证据层）

Demo type name: `IntelEntry` / `IntelDetail` in `webui/src/types/intel.ts`.

| Field | Type | Required | Notes |
|-------|------|----------|-------|
| `id` | int | PK | |
| `title` | string | yes | 报告原标题 |
| `source_id` | int | FK | → Source |
| `source_url` | string | no | |
| `disclosed_at` | datetime | no | |
| `collected_at` | datetime | yes | |
| `original_body` | text | no | HTML/MD 快照 |
| `ingest_status` | enum | draft / parsing / pending_review / published / archived |
| `parse_version` | string | no | |
| `attack_knowledge_id` | int | FK | Phase 2; Phase 1 implicit 1:1 |

**Not on IntelReport (belong to AttackKnowledge):** `summary` (knowledge), `attack_matrix`, `target_industries`, primary TTP bindings — unless Phase 1 colocated for demo.

---

## knowledge_intel_link

| Field | Type | Notes |
|-------|------|-------|
| `knowledge_id` | int | FK AttackKnowledge |
| `intel_id` | int | FK IntelReport |
| `role` | enum | primary / supporting |
| `linked_at` | datetime | |

---

## ThreatActor / Malware / MitreTechnique / Procedure / Ioc

- **Procedure**, **KnowledgeTechnique**: FK → `attack_knowledge_id`
- **Ioc**: FK → `intel_report_id` (primary); optional link to knowledge via report
- **ThreatActor / Malware**: M:N with `AttackKnowledge` (`involves`)

See previous sections in git history for field lists; unchanged except FK targets.

---

## STIX Export (Phase 2)

- Export **AttackKnowledge** as custom extension or aggregated `grouping` + `attack-pattern` relationships
- Export **IntelReport** as STIX `report` with `object_refs` pointing to observables and patterns
- Do **not** treat AMI as indicator-first STIX feed
