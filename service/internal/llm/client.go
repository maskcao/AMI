package llm

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/ami/ami/service/internal/store"
)

type MalwareEntry struct {
	Name           string `json:"name"`
	Classification string `json:"classification"`
}

type InfraEntry struct {
	Type  string `json:"type"`
	Value string `json:"value"`
}

type ToolEntry struct {
	Name string `json:"name"`
	Role string `json:"role,omitempty"`
}

type EnrichResult struct {
	Summary             string
	SuggestedCves       []string
	SuggestedTechniques []string
	SuggestedMalware    []string
	Tools               []ToolEntry
	Infrastructure      []InfraEntry
	Malware             []MalwareEntry
	Quality             QualityReport
}

type Client struct {
	BaseURL      string
	Model        string
	APIKey       string
	ProviderID   string
	ProviderName string
	HTTP         *http.Client
}

func NewFromDB(db *store.DB) *Client {
	providers, err := orderedEnrichProviders(db)
	if err != nil || len(providers) == 0 {
		p, _ := db.DefaultLLMProvider()
		if p == nil {
			return nil
		}
		return ClientFromProvider(p)
	}
	c := ClientFromProvider(&providers[0])
	return c
}

func (c *Client) Enabled() bool {
	return c != nil && c.APIKey != "" && c.BaseURL != ""
}

type chatReq struct {
	Model    string        `json:"model"`
	Messages []chatMessage `json:"messages"`
}

type chatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type chatResp struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
}

func (c *Client) Enrich(ctx context.Context, text string) (*EnrichResult, error) {
	if !c.Enabled() {
		return nil, fmt.Errorf("llm client not configured")
	}
	const maxChars = 10000
	if len(text) > maxChars {
		text = text[:maxChars]
	}
	systemPrompt := `You are a threat-intelligence extraction engine. Each request is ONE standalone article; ignore any prior context.
Read only the user article text. Respond with a single JSON object (no markdown fences).

Required schema:
{
  "summary": "2-4 sentences in Chinese describing the attack event (who/what/how/impact). No JSON-LD, no copyright boilerplate.",
  "cves": ["CVE-YYYY-NNNN"],
  "techniques": ["T1234"],
  "malware": [{"name":"...", "classification":"botnet|ransomware|backdoor|trojan|worm|miner|apt|other"}],
  "tools": [{"name":"...", "role":"optional short role"}],
  "infrastructure": [{"type":"domain|ip|url|c2", "value":"..."}]
}

Rules:
- Use empty arrays when unknown; do not invent CVEs or technique IDs.
- techniques must be MITRE ATT&CK IDs like T1059 or T1059.001.
- malware.classification must be one of the enum values above.
- infrastructure.value must be concrete indicators from the text.
- Do not repeat the article title alone as summary.`

	body, _ := json.Marshal(chatReq{
		Model: c.Model,
		Messages: []chatMessage{
			{Role: "system", Content: systemPrompt},
			{Role: "user", Content: text},
		},
	})
	if hint := validateVolcengineModel(c.BaseURL, c.Model); isVolcengineBase(c.BaseURL) && hint != "" {
		return nil, fmt.Errorf("volcengine config: %s", hint)
	}
	url := chatCompletionsURL(c.BaseURL)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.APIKey)
	if isVolcengineBase(c.BaseURL) {
		applyVolcengineHeaders(req)
	}

	httpClient := c.HTTP
	if httpClient == nil {
		httpClient = &http.Client{Timeout: 90 * time.Second}
	}
	res, err := httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	b, _ := io.ReadAll(res.Body)
	if res.StatusCode >= 400 {
		msg := truncate(string(b), 200)
		if isVolcengineBase(c.BaseURL) {
			msg = friendlyVolcengineError(res.StatusCode, string(b), c.BaseURL, c.Model)
		}
		return nil, fmt.Errorf("llm http %d: %s", res.StatusCode, msg)
	}
	var cr chatResp
	if err := json.Unmarshal(b, &cr); err != nil {
		return nil, err
	}
	if len(cr.Choices) == 0 {
		return nil, fmt.Errorf("llm empty response")
	}
	content := strings.TrimSpace(cr.Choices[0].Message.Content)
	content = strings.TrimPrefix(content, "```json")
	content = strings.TrimPrefix(content, "```")
	content = strings.TrimSuffix(content, "```")
	content = strings.TrimSpace(content)

	var raw struct {
		Summary          string         `json:"summary"`
		Cves             []string       `json:"cves"`
		Techniques       []string       `json:"techniques"`
		Malware          []MalwareEntry `json:"malware"`
		MalwareLegacy    []string       `json:"malware_names"`
		Tools            []ToolEntry    `json:"tools"`
		Infrastructure   []InfraEntry   `json:"infrastructure"`
	}
	if err := json.Unmarshal([]byte(content), &raw); err != nil {
		return nil, fmt.Errorf("llm json parse: %w", err)
	}
	malwareNames := raw.MalwareLegacy
	for _, m := range raw.Malware {
		if strings.TrimSpace(m.Name) != "" {
			malwareNames = append(malwareNames, m.Name)
		}
	}
	er := &EnrichResult{
		Summary:             strings.TrimSpace(raw.Summary),
		SuggestedCves:       raw.Cves,
		SuggestedTechniques: raw.Techniques,
		SuggestedMalware:    malwareNames,
		Tools:               raw.Tools,
		Infrastructure:      raw.Infrastructure,
		Malware:             raw.Malware,
	}
	er.Quality = ValidateEnrich(er)
	return er, nil
}

func normalizeBaseURL(p *store.ModelProviderRow) string {
	if !p.BaseURL.Valid || p.BaseURL.String == "" {
		return ""
	}
	base := strings.TrimRight(strings.TrimSpace(p.BaseURL.String), "/")
	if isVolcengineBase(base) {
		switch {
		case strings.Contains(base, "/api/coding"):
			// keep Hermes coding plan URL as-is
		case strings.HasSuffix(base, "/api/plan"):
			base = ArkInferenceBase
		case !strings.Contains(base, "/api/v3") && !strings.HasSuffix(base, "/v1"):
			if strings.HasSuffix(base, "volces.com") || strings.Contains(base, "ark.cn-beijing") {
				base = ArkInferenceBase
			}
		}
	}
	return base
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "..."
}
