# UI Field Specification

> **Demo routes:** Home `/knowledge` = attack knowledge; `/intel` = intel workbench; `/ioc` = auxiliary IOC search.

## List (`/knowledge`) — 攻击知识库

| Column | Field |
|--------|-------|
| ID | `id` |
| 攻击知识 | `title` |
| 矩阵 | `attackMatrix` |
| 攻击行业 | `targetIndustries` |
| 恶意软件 | `malwareFamilies` |
| 支撑情报 | `intelCount` |
| 质量评分 | `qualityScore` |

## Detail (`/knowledge/:id`)

Tabs: **知识分析** (summary, ATT&CK, entities) | **支撑情报** (`linkedIntel[]`)

## List (`/intel`) — 情报工作台

| Column | Field |
|--------|-------|
| ID | `id` |
| 情报标题 | `title` |
| 披露时间 | `disclosedAt` |
| 来源 | `sourceName` |
| 状态 | `ingestStatus` |
| 关联攻击知识 | `knowledgeIds[]` |
| IOC | `iocCount` |
| 质量评分 | `qualityScore` |

## Detail (`/intel/:id`)

Tabs: **关联攻击知识** | **IOC（附属）** | **原始信息**

## IOC (`/ioc`)

Search → hit intel + linked `knowledgeIds`
