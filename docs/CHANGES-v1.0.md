# AMI 系统变更清单（相对最初系统）

分支：`dev/ami-v1.0`，基线为初始提交（`267a12b8 init project` / `ff00f474`）之后的全部修改。

## 一、新功能

### 1. 安全事件报告模块
- 后端：
  - 新表 `incident_report`（[service/internal/store/sql/004_incident_reports.sql](../service/internal/store/sql/004_incident_reports.sql)），含标题、事件时间、报告时间、严重级别、附件文件等字段；`report_time` 列由 [schema_ensure.go](../service/internal/store/schema_ensure.go) 自动补列。
  - 存取层 [service/internal/store/incident.go](../service/internal/store/incident.go)：创建、列表（关键词/级别过滤 + 分页）、详情、更新、删除、附件登记。
  - API [service/internal/api/incident.go](../service/internal/api/incident.go)：`/api/v1/incident-reports` CRUD + 附件上传/下载/预览。
- 前端：
  - 页面 [webui/src/views/datacenter/IncidentReportPage.vue](../webui/src/views/datacenter/IncidentReportPage.vue)：搜索 + 上传工具栏（FilterCard 方框包裹，与其他模块一致）、表格分页、上传后可预览/下载。
  - API 封装 [webui/src/api/incident.ts](../webui/src/api/incident.ts)；路由 `/datacenter/reports`（[router/index.ts](../webui/src/router/index.ts)）；侧边栏"安全事件报告"入口（[AppSidebar.vue](../webui/src/components/AppSidebar.vue)）。
  - 依赖新增：`mammoth`（docx 预览）、`marked`（markdown 渲染）。

### 2. LLM 降级标记 + 批量重新抽取
- 背景：LLM 富集失败时流水线回退到规则抽取（`enrich_version = 'rules-v1'`），此前无法区分、无法重跑。
- 后端：
  - [store/knowledge.go](../service/internal/store/knowledge.go)：`KnowledgeRow` 增加 `EnrichVersion`；列表 SQL 带出最新关联情报的 `enrich_version`；新增 `llmFailed=true` 过滤（有关联情报但全部不是 `llm-v2*`）。
  - [store/llm_retry.go](../service/internal/store/llm_retry.go)：`LLMFailedIntel` 查询待重跑情报 + `RequeueLLMFailed` 清理旧任务并重新入队 `extract` 阶段。
  - API：`POST /api/v1/intel/llm-retry?limit=N`（[intel_retry.go](../service/internal/api/intel_retry.go)，单次上限 500、默认 200，每条消耗一次 LLM 调用）；知识列表/详情返回 `enrichVersion`。
- 前端：
  - 知识列表行内红色"LLM 失败"标签（[KnowledgeCards.vue](../webui/src/components/KnowledgeCards.vue)）。
  - 复核范围新增"LLM 失败"筛选项；页头"重跑 LLM 失败"按钮，带确认弹窗（[KnowledgeList.vue](../webui/src/views/KnowledgeList.vue)、[api/knowledge.ts](../webui/src/api/knowledge.ts)）。

### 3. 疑似重复知识检测
- 后端：
  - [store/knowledge_dup.go](../service/internal/store/knowledge_dup.go)（新建）：标题 token 相似度（Jaccard + 包含率）+ CVE / ATT&CK 手法 / 恶意软件三维重叠加权评分（阈值 0.45，取 Top 5）；入库（`CreateKnowledge` / `UpdateKnowledgeSummary`）时自动计算并存入新列 `dup_candidates_json`（由 [schema_ensure.go](../service/internal/store/schema_ensure.go) 自动补列）。
  - 一次性回扫端点 `POST /api/v1/knowledge/dup-scan`（[knowledge_dup.go](../service/internal/api/knowledge_dup.go)），全量重算存量数据（当前库 501 条被标记）。
  - 合并知识后自动重算目标条目的重复候选（[knowledge_patch.go](../service/internal/store/knowledge_patch.go)）。
- 前端：
  - 知识列表"疑似重复"红色标签；知识详情页"疑似重复"卡片，展示相似候选（相似度、重叠原因），支持一键"并入该条"合并去重（[KnowledgeDetail.vue](../webui/src/views/KnowledgeDetail.vue)）。

## 二、缺陷修复

### 4. 威胁情报库实体详情"未找到实体"
- 根因：`getEntityProfile` 只查 mock 数据，API 模式下永远找不到实体。
- 修复：[webui/src/analytics/entity-profiles.ts](../webui/src/analytics/entity-profiles.ts) 新增 `getEntityProfileAsync`——API 模式下从 analytics bundle 定位实体行，按关联知识逐条拉取详情后聚合构建档案；[EntityDimensionDetailPage.vue](../webui/src/views/EntityDimensionDetailPage.vue) 改用异步加载。覆盖漏洞利用 / 恶意软件 / 攻击组织 / 基础设施等全部维度。

### 5. 情报下载修复
- [service/internal/artifacts/store.go](../service/internal/artifacts/store.go)：修正附件/产物读取路径解析，上传与预览/下载可端到端使用。

### 6. SQLite 并发稳定性
- [service/internal/store/db.go](../service/internal/store/db.go)：启用 WAL 模式并设置 `busy_timeout=5000`，消除并发读写时偶发 `database is locked`。

## 三、优化

### 7. Analytics Bundle 进程内缓存
- [service/internal/analytics/cache.go](../service/internal/analytics/cache.go)（新建）：按数据水位（`COUNT + MAX(updated_at)` over `attack_knowledge` 与 `intel_report`）失效的进程内缓存，数据变更自动重建。
- [service/internal/api/analytics.go](../service/internal/api/analytics.go)：`/analytics/bundle` 等改用缓存。实测首次 ~0.96s → 命中 ~0.07s（约 14 倍）。

### 8. CVE 维度按披露时间倒序
- [service/internal/analytics/build.go](../service/internal/analytics/build.go)：仅漏洞利用（CVE）维度改为按披露时间倒序（无披露时间的回退到更新时间，平局按名称），MITRE 手法及其他维度保持原字典序不变。

### 9. 攻击知识库筛选区 UI
- [webui/src/views/KnowledgeList.vue](../webui/src/views/KnowledgeList.vue)：筛选条件单行 flex 布局（关键词/攻击行业弹性伸缩，下拉固定宽度，按钮不换行），标签右对齐不换行，重置按钮可清空全部条件；窄屏（≤640px）标签上置。

### 10. 仓库瘦身
- 新增 `.gitignore` 排除运行时数据目录（`ff00f474`）；工作区重复文件硬链接去重，磁盘占用减少约 8.94 GB。

## 四、验证要点
- `cd service && go build ./...` 通过；`go vet` 无新增告警。
- `cd webui && npx vue-tsc --noEmit`：本次触碰文件无类型错误（仓库存在预存历史错误，与本次无关）。
- 浏览器实测：知识列表三种标签（待复核 / LLM 失败 / 疑似重复）、"LLM 失败"筛选、重跑确认弹窗、详情页疑似重复卡片均正常；`POST /intel/llm-retry?limit=1` 重跑后日志出现 `llm enrich ok`。
