# 局域网访问说明（192 网段）

本机示例 IP：**192.168.238.10**（以 `hostname -I` / `ip addr` 为准）。

## 启动（监听所有网卡）

```bash
# 终端 1：API（0.0.0.0:8080）
cd service
AMI_ADDR=0.0.0.0:8080 AMI_DATA_DIR=./data AMI_SEED=1 go run ./cmd/ami -seed

# 终端 2：WebUI（0.0.0.0:5173，/api 代理到本机 8080）
cd webui
cp .env.development.example .env.development   # VITE_USE_API=true
npm run dev
```

防火墙需放行：**5173**（WebUI）、**8080**（API，仅直连调试时需要）。

## 业务页面（WebUI）

将 `192.168.238.10` 换成你的机器 IP。

| 功能 | 地址 |
|------|------|
| 首页（运营总览） | http://192.168.238.10:5173/dashboard |
| 攻击知识库 | http://192.168.238.10:5173/knowledge |
| 知识详情 | http://192.168.238.10:5173/knowledge/1001 |
| 采集队列 | http://192.168.238.10:5173/intel |
| 运行日志 | http://192.168.238.10:5173/datacenter/logs |
| 情报详情 | http://192.168.238.10:5173/intel/4203 |
| IOC 检索 | http://192.168.238.10:5173/ioc |
| 威胁库 · CVE | http://192.168.238.10:5173/analytics/cve |
| 威胁库 · 恶意软件 | http://192.168.238.10:5173/analytics/malware |
| 威胁库 · 攻击组织 | http://192.168.238.10:5173/analytics/actors |
| 威胁库 · MITRE | http://192.168.238.10:5173/analytics/mitre |
| 威胁库 · 工具 | http://192.168.238.10:5173/analytics/tools |
| 威胁库 · 基础设施 | http://192.168.238.10:5173/analytics/infra |
| 系统配置 · 采集 | http://192.168.238.10:5173/settings/collect |
| 系统配置 · 解析模板 | http://192.168.238.10:5173/settings/parser |
| 系统配置 · 模型服务 | http://192.168.238.10:5173/settings/provider |

浏览器访问 WebUI 时，API 走同源代理：`http://192.168.238.10:5173/api/v1/...` → 本机 `127.0.0.1:8080`。

## REST API（直连，Postman/curl）

基址：**http://192.168.238.10:8080**

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | `/health` | 健康检查 |
| GET | `/api/v1/knowledge` | 攻击知识列表 |
| GET | `/api/v1/knowledge/{id}` | 知识详情 |
| PATCH | `/api/v1/knowledge/{id}` | 更新知识 |
| POST | `/api/v1/knowledge/{id}/publish` | 发布 |
| POST | `/api/v1/knowledge/merge` | 合并知识 |
| GET | `/api/v1/intel` | 采集队列列表 |
| GET | `/api/v1/intel/{id}` | 情报详情（默认不含完整原文） |
| GET | `/api/v1/intel/{id}/original` | 原文/物证（按需） |
| POST | `/api/v1/intel/import` | 导入 HTML |
| POST | `/api/v1/intel/{id}/retry` | 重试流水线 |
| GET | `/api/v1/ioc?q=` | IOC 检索 |
| GET | `/api/v1/analytics/bundle` | 运营总览 + 威胁库聚合 |
| GET | `/api/v1/analytics/actors/{slug}` | 攻击组织档案 |
| GET | `/api/v1/settings/collect-sources` | 采集源列表 |
| GET | `/api/v1/settings/model-providers` | 模型服务列表 |
| GET | `/api/v1/settings/parser-profiles` | 解析模板列表 |
| GET | `/api/v1/datacenter/runtime-logs` | 运行日志（`level` / `source` / `q` / `limit`） |
| POST | `/api/v1/datacenter/runtime-logs/clear` | 清空运行日志 |

示例：

```bash
curl http://192.168.238.10:8080/health
curl http://192.168.238.10:8080/api/v1/knowledge
```
