# 火山方舟（Volcengine Ark）与 AMI / Hermes 对齐说明

## 现象

- Hermes 能调通，AMI「模型服务 → 测试」或流水线 `llm enrich` 报 **401 AuthenticationError**。
- 常见原因：**API Key 未同步到 AMI 数据库**，或 **Base URL / 模型 ID 与 Hermes 不一致**。

## 两套 OpenAI 兼容地址（不要混用）

| 用途 | Base URL | `model` 字段（接入点/模型名） |
|------|----------|-------------------------------|
| **在线推理**（情报抽取推荐） | `https://ark.cn-beijing.volces.com/api/v3` | 控制台「推理接入点 ID」：`ep-…` 或 `ark-…` |
| **编程计划**（Hermes 部分场景） | `https://ark.cn-beijing.volces.com/api/coding/v3` | Hermes 配置里的 `model`（如 `ark-code-latest`） |

勿使用 `https://ark.cn-beijing.volces.com/api/plan`。

## 从 Hermes 同步到 AMI

1. 打开 Hermes 配置（如 `~/.hermes/config.yaml`）中找到 `providers.ark`（名称可能不同）。
2. 在 AMI **系统配置 → 模型服务** 新建或编辑一条：
   - **Base URL**：与 Hermes 的 `url` **完全一致**（含 `/api/v3` 或 `/api/coding/v3`）。
   - **接入点 / 模型 ID**：与 Hermes 的 `model` **完全一致**。
   - **API Key**：与 Hermes 的 `api_key` **完全一致**（方舟「大模型 API Key」，不是云账号 AK/SK）。
3. 点击 **测试**，显示连通后再跑采集队列 **重试流水线**。

可选：在运行 API 的机器上设置环境变量（优先于库内 Key，仅方舟域名生效）：

```bash
export AMI_ARK_API_KEY="你的方舟APIKey"
```

## AMI 已做的兼容

- 请求头 `User-Agent: curl/8.0`（与 Hermes 社区方案一致）。
- 按模型列表依次尝试已启用配置（`EnrichWithDB`）。
- 401 / UnsupportedModel 时返回可读的配置提示。

## 本环境实测（供排查）

- 库内旧 Key 对 `/api/v3`：**401**（Key 对方舟推理无效或已过期）。
- 同一 Key 对 `/api/coding/v3` + 推理接入点 ID：**UnsupportedModel**（说明 Key 能到达网关，但 URL 与模型类型不匹配）。

**结论**：请用 Hermes 里**当前可用**的三项（URL、model、api_key）覆盖写入 AMI，不要只改 URL 不改 Key。
