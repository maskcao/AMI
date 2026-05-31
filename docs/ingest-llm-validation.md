# 采集源 LLM 解析 / 抽取技术验证

基于当前系统配置采集源（示例：`360netlab-botnet`，类型 `scheduled_url`，列表 `https://blog.netlab.360.com/tag/botnet/`）的代码与运行时证据。

## 1. Prompt 与会话窗口

| 项 | 结论 | 证据 |
|----|------|------|
| 每次解析独立 | **是** | `llm/client.go` 每次 `Enrich` 仅 `system` + `user` 两条消息，无历史 |
| 上下文截断 | **是** | `body.md` 送入 LLM 前截断至 10000 字符（`llm.Client.Enrich`） |
| Prompt 清晰度 | **已加强（llm-v2）** | 明确要求 overview / CVE / ATT&CK / malware+classification / tools / infrastructure；禁止 JSON-LD 进 summary |
| 无意义正文 | **已缓解** | `pipeline/textprep.go` 剥离 script、JSON-LD、schema.org 大块 |

## 2. 质量检查维度

| 维度 | 实现 | 说明 |
|------|------|------|
| （1）攻击事件概述 | `llm.ValidateEnrich` | 长度、禁止 schema.org 噪声 |
| （2）攻击图示 | `SyncFiguresFromHTML` | 解析阶段下载至 `parsed/v1/figures/`，API `GET /api/v1/intel/{id}/artifacts/{name}` |
| （3）恶意软件 / ATT&CK / 工具 / 基础设施 / CVE | `EnrichResult` + `applyEnrichToPayload` | 写入 `attack_knowledge.payload_json` |
| （4）恶意软件分类 | `malware[].classification` | 枚举：botnet、ransomware、backdoor 等 |

质量结果暴露在情报详情 `aiExtract.qualityHints` / `qualityOk`。

## 3. 运行时验证（2026-05-31）

- 采集源：`GET /api/v1/settings/collect-sources` → 1 条启用源 `360netlab-botnet`
- 图示表：`intel_figure` 在改动前为 **0 行**（流水线未落库）
- 概述：`enrich_version=rules-v1` 时 summary 含 **schema.org JSON**（旧版 `body.txt` 或未重跑 `html-v2`）
- LLM：`AMI_LLM_INTEGRATION=1 go test ./internal/llm/...` → **HTTP 401**，DB 内 API Key 无效；流水线回退 rules

### 建议运维步骤

1. 在「系统配置 → 模型提供方」更新有效 API Key（或设置 `AMI_LLM_API_KEY`）
2. `cd service && go build ./... &&` 重启 API
3. 对样本情报执行 `POST /api/v1/intel/{id}/retry`
4. 检查：`enrich_version=llm-v2`、`ai_extract_json.qualityOk`、`intel_figure` 行数、知识库 payload 各数组

## 4. 自动化测试

```bash
cd service
go test ./internal/pipeline/... ./internal/llm/... -count=1
# 可选真实 LLM（需有效 Key）：
AMI_LLM_INTEGRATION=1 go test ./internal/llm/... -run TestEnrichIntegration -v
```
