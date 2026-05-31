# 采集物证双份留存与 LLM 输入说明

## 1. 两份本地留存（parse 阶段 `html-v2`）

| 层级 | 路径 | 内容 |
|------|------|------|
| **L0 原始 HTML + 资源** | `store/intel/{id}/raw/001.html` | 采集页 HTML；`ArchivePageAssets` 将文内 **img / link[stylesheet] / script[src]** 下载到 `raw/001/assets/`，并把引用改写为 `/api/v1/intel/{id}/raw-assets/{hash}.ext` |
| **L1 Markdown（含 LLM 输入）** | `parsed/v1/body.md` | `HTMLToMarkdown` 由 HTML 转换；**extract 阶段直接读取此文件**，不再生成 `body.txt` |

图示（报告插图）另存：`parsed/v1/figures/`（`intel_figure` 表，供知识库/采集详情画廊）。

### API 调阅

| 接口 | 用途 |
|------|------|
| `GET /api/v1/intel/{id}/original` | 本地归档 HTML（含已改写资源 URL） |
| `GET /api/v1/intel/{id}/raw-assets/{name}` | 单个子资源（css/js/图片） |
| `GET /api/v1/intel/{id}/markdown` | Markdown 正文 |
| `GET /api/v1/intel/{id}/parsed` | 与 markdown 相同（body.md，兼容旧客户端） |

**说明**：外链 JS 仅下载落盘，浏览器内嵌执行仍受 CSP/沙箱限制；离线阅读以 HTML+CSS+图片为主。

## 2. LLM 当前分析的是哪部分数据

流水线 **`extract` 阶段**（`worker.doExtract`）读取顺序：

1. `intel_parsed.body_md_path` → `parsed/v1/body.md`（**首选**）
2. 若为空 → 遗留 `body_txt_path` / `body.txt`（重跑前旧数据）
3. 仍为空 → `intel_report.body_preview`（采集时 2KB 预览）

**不直接读**：原始 HTML 二进制、图片二进制。图片信息仅当 `body.md` 中仍含 alt/链接文字时间接进入模型。

截断：送入模型前最多 **10000 字符**（`llm.Client.Enrich`）。

模型来源：`EnrichWithDB` 按「模型服务」列表依次调用，成功则 `enrich_version=llm-v2:{providerId}`，否则回退 `rules-v1`（规则摘要 + CVE 正则）。

## 3. LLM 提示词格式（`llm-v2`）

**调用方式**：OpenAI 兼容 `POST {baseUrl}/chat/completions`，**无会话历史**（仅 2 条 message）。

**system**（摘录，完整见 `service/internal/llm/client.go`）：

```text
You are a threat-intelligence extraction engine. Each request is ONE standalone article...
Required schema:
{
  "summary": "2-4 sentences in Chinese ...",
  "cves": ["CVE-YYYY-NNNN"],
  "techniques": ["T1234"],
  "malware": [{"name":"...", "classification":"botnet|ransomware|..."}],
  "tools": [{"name":"...", "role":"..."}],
  "infrastructure": [{"type":"domain|ip|url|c2", "value":"..."}]
}
Rules: empty arrays when unknown; techniques must be MITRE IDs; ...
```

**user**：`body.md` 全文（≤10000 字符）。

**期望响应**：单个 JSON 对象（无 markdown 代码块包裹）；解析后写入 `intel_report.ai_extract_json` 并映射到 `attack_knowledge.payload_json`。

## 4. 重跑以生成新留存

对已有情报需 **重试流水线**（或重新采集）后才会生成 `html-v2` 归档资源；重跑后不再更新 `body.txt`。
