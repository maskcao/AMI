# UI Field Specification

> **Demo routes:** Home `/knowledge` = attack knowledge; `/intel` = intel ingest queue; `/ioc` = auxiliary IOC search.

## List (`/knowledge`) — 攻击知识库

| Column | Field |
|--------|-------|
| ID | `id` |
| 标题 | `title` + `needsReview` badge |
| 摘要 | `summary` (truncated) |
| 状态 | `status` (`draft` / `published`) |
| 矩阵 | `attackMatrix` |
| 攻击行业 | `targetIndustries` |
| 恶意软件 | `malware[]` |
| 知识维度 | `dimensions.*` counts |
| 支撑情报 | `intelCount` |
| 质量评分 | `qualityScore` |
| 更新时间 | `updatedAt` (default sort desc) |

Filters: keyword (`title` / `summary` / CVE / Technique ID / actors / malware), `review` dropdown (`all` \| `needs_review` \| `draft` \| `published`), industry, matrix, malware class. Header shows `needsReviewCount`. Table header sort: `updatedAt` (default desc), `id`, `intelCount`, `qualityScore` — no separate sort control in filter bar.

## Detail (`/knowledge/:id`)

- **流水线溯源** (`pipeline`): last intel, `knowledgeAction`, `extractConfidence`, link to `/intel/:id` and 采集队列
- **可编辑保存**: `summary`, `targetIndustries` → `updateKnowledge` (Mock in-memory)
- **发布**: `publishKnowledge` (`draft` → `published`)
- **合并** (Demo): `mergeKnowledge` — attach `linkedIntel` to target, remove source

Tabs: **攻击维度** (entities + `procedures[]`) | **关联视图** | **支撑情报** (`linkedIntel[]`)

## List (`/intel`) — 采集队列

| Column | Field |
|--------|-------|
| ID | `id` |
| 标题 | `title` |
| 采集路径 | `sourceUrl`（原文 URL，可外链打开） |
| 采集时间 | `collectedAt` |
| 处理状态 | `ingestStatus` |
| 知识写入 | `knowledgeAction` |
| 关联攻击知识 | `knowledgeIds[]` |
| 抽取置信 | `extractConfidence` |

## Detail (`/intel/:id`)

Tabs: **流水线** | **AI 抽取** | **攻击知识（自动）** | **IOC（附属）** | **原始信息**

## List (`/analytics/cve`) — 漏洞利用

| Column | Field |
|--------|-------|
| CVE | `cveId` → `/analytics/entity/cve/:entityKey` |
| 描述 | `description` |
| 在野 | `inWild` |
| 攻击行业 | `targetIndustries[]` |
| 关联知识 | `knowledgeCount` |

## Detail (`/analytics/entity/:dimension/:entityKey`)

`dimension`: `cve` \| `malware` \| `tool` \| `infra` — Tabs: 关联图谱 + 各维度明细（与攻击组织详情结构一致）

## List (`/analytics/actors`) — 攻击组织

| Column | Field |
|--------|-------|
| 组织/团伙 | `name` → link `/analytics/actor/:actorKey` |
| 攻击行业 | `targetIndustries[]` (aggregated) |
| 关联知识 | `knowledgeCount` |

## Detail (`/analytics/actor/:actorKey`)

Tabs: **关联图谱** | **攻击知识** | **攻击手法** | **攻击工具** | **基础设施** | **恶意软件** | **利用漏洞**

## IOC (`/ioc`)

Search → hit intel + linked `knowledgeIds`
