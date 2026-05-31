package llm

import (
	"net/http"
	"os"
	"strings"

	"github.com/ami/ami/service/internal/store"
)

// Volcengine Ark OpenAI-compatible bases (see Hermes / 方舟控制台).
const (
	ArkInferenceBase = "https://ark.cn-beijing.volces.com/api/v3"
	ArkCodingBase    = "https://ark.cn-beijing.volces.com/api/coding/v3"
)

func isVolcengineBase(base string) bool {
	return strings.Contains(strings.ToLower(base), "volces.com")
}

func chatCompletionsURL(base string) string {
	base = strings.TrimRight(strings.TrimSpace(base), "/")
	if strings.HasSuffix(base, "/chat/completions") {
		return base
	}
	return base + "/chat/completions"
}

func applyVolcengineHeaders(req *http.Request) {
	// Hermes workaround: Ark may reject default OpenAI SDK User-Agent.
	req.Header.Set("User-Agent", "curl/8.0")
}

func arkAPIKey(p *store.ModelProviderRow) string {
	if k := strings.TrimSpace(os.Getenv("AMI_ARK_API_KEY")); k != "" {
		k = strings.TrimPrefix(k, "Bearer ")
		k = strings.TrimPrefix(k, "bearer ")
		return strings.TrimSpace(k)
	}
	return resolveAPIKey(p)
}

func volcengineKind(base string) string {
	low := strings.ToLower(base)
	if strings.Contains(low, "/api/coding/") {
		return "coding"
	}
	if strings.Contains(low, "/api/plan") {
		return "plan"
	}
	return "inference"
}

func validateVolcengineModel(base, model string) string {
	model = strings.TrimSpace(model)
	if model == "" {
		return "接入点 / 模型 ID 不能为空"
	}
	kind := volcengineKind(base)
	if kind == "plan" {
		return "Base URL 请勿使用 /api/plan；推理用 " + ArkInferenceBase + "，Hermes 编程计划用 " + ArkCodingBase
	}
	if strings.EqualFold(model, "auto") {
		return "请填写控制台「推理接入点 ID」（ep- 或 ark- 开头），不要使用 auto"
	}
	if kind == "coding" && (strings.HasPrefix(model, "ep-") || strings.Contains(model, "6156f9a5")) {
		return "编程计划 URL 下模型名需与 Hermes 一致（如 ark-code-latest）；推理接入点请改用 " + ArkInferenceBase
	}
	return ""
}

func friendlyVolcengineError(status int, body, base, model string) string {
	body = truncate(body, 280)
	if hint := validateVolcengineModel(base, model); hint != "" && (status == 400 || status == 401) {
		return hint + " | " + body
	}
	switch {
	case status == 401 || strings.Contains(body, "AuthenticationError"):
		return "API Key 无效：请在「模型服务」填入与 Hermes 相同的方舟 Key，或设置 AMI_ARK_API_KEY 后重启 API | " + body
	case strings.Contains(body, "UnsupportedModel"):
		return "模型与 Base URL 类型不匹配（推理 /api/v3 ↔ 编程 /api/coding/v3）| " + body
	default:
		return body
	}
}
