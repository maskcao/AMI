package llm

import (
	"context"
	"fmt"
	"log"
	"os"
	"sort"
	"strings"

	"github.com/ami/ami/service/internal/store"
)

// EnrichOutcome is the result of trying configured model providers.
type EnrichOutcome struct {
	Result     *EnrichResult
	VersionTag string // e.g. llm-v2:mp-xxx
	ProviderID string
	Provider   string
}

// ClientFromProvider builds an HTTP client from a model_provider row.
func ClientFromProvider(p *store.ModelProviderRow) *Client {
	if p == nil || !p.Enabled || p.Kind != "llm" {
		return nil
	}
	base := normalizeBaseURL(p)
	key := resolveAPIKey(p)
	if isVolcengineBase(base) {
		key = arkAPIKey(p)
	}
	if key == "" {
		return nil
	}
	if base == "" {
		base = "https://api.openai.com/v1"
	}
	model := strings.TrimSpace(p.ModelID.String)
	if model == "" {
		model = "gpt-4o-mini"
	}
	return &Client{
		BaseURL: base,
		Model:   model,
		APIKey:  key,
		ProviderID: p.ID,
		ProviderName: p.Name,
	}
}

func resolveAPIKey(p *store.ModelProviderRow) string {
	if p.APIKey.Valid {
		k := strings.TrimSpace(p.APIKey.String)
		k = strings.TrimPrefix(k, "Bearer ")
		k = strings.TrimPrefix(k, "bearer ")
		if k != "" {
			return k
		}
	}
	if k := os.Getenv("AMI_LLM_API_KEY"); k != "" {
		return strings.TrimSpace(k)
	}
	return strings.TrimSpace(os.Getenv("OPENAI_API_KEY"))
}

func providerSortKey(p store.ModelProviderRow) int {
	score := 0
	if p.DefaultFor.Valid && strings.Contains(p.DefaultFor.String, "enrich") {
		score += 10
	}
	mid := ""
	if p.ModelID.Valid {
		mid = strings.TrimSpace(p.ModelID.String)
	}
	switch {
	case strings.HasPrefix(mid, "ep-"):
		score += 30
	case strings.HasPrefix(mid, "ark-") && len(mid) > 12:
		score += 25
	case mid != "" && !strings.EqualFold(mid, "auto"):
		score += 5
	case strings.EqualFold(mid, "auto"):
		score -= 40
	}
	if !p.APIKey.Valid || strings.TrimSpace(p.APIKey.String) == "" {
		score -= 100
	}
	return score
}

func orderedEnrichProviders(db *store.DB) ([]store.ModelProviderRow, error) {
	all, err := db.ListModelProviders()
	if err != nil {
		return nil, err
	}
	var candidates []store.ModelProviderRow
	for _, p := range all {
		if !p.Enabled || p.Kind != "llm" {
			continue
		}
		if resolveAPIKey(&p) == "" {
			continue
		}
		if p.ModelID.Valid {
			mid := strings.TrimSpace(p.ModelID.String)
			if mid == "" {
				continue
			}
		} else {
			continue
		}
		candidates = append(candidates, p)
	}
	sort.Slice(candidates, func(i, j int) bool {
		return providerSortKey(candidates[i]) > providerSortKey(candidates[j])
	})
	return candidates, nil
}

// EnrichWithDB tries each enabled model in settings until one succeeds.
func EnrichWithDB(db *store.DB, ctx context.Context, text string) (*EnrichOutcome, error) {
	providers, err := orderedEnrichProviders(db)
	if err != nil {
		return nil, err
	}
	if len(providers) == 0 {
		return nil, fmt.Errorf("no enabled model provider with api key and model id")
	}
	var errs []string
	for _, p := range providers {
		c := ClientFromProvider(&p)
		if c == nil || !c.Enabled() {
			continue
		}
		er, err := c.Enrich(ctx, text)
		if err == nil && er != nil {
			tag := fmt.Sprintf("llm-v2:%s", p.ID)
			log.Printf("llm enrich ok provider=%s model=%s", p.Name, c.Model)
			return &EnrichOutcome{
				Result:     er,
				VersionTag: tag,
				ProviderID: p.ID,
				Provider:   p.Name,
			}, nil
		}
		msg := fmt.Sprintf("%s(%s): %v", p.Name, p.ID, err)
		errs = append(errs, msg)
		log.Printf("llm enrich try %s", msg)
	}
	return nil, fmt.Errorf("all model providers failed: %s", strings.Join(errs, "; "))
}

// TestProvider pings a single model_provider with a minimal completion.
func TestProvider(ctx context.Context, p *store.ModelProviderRow) error {
	c := ClientFromProvider(p)
	if c == nil || !c.Enabled() {
		return fmt.Errorf("provider not configured (missing api key, base url, or model id)")
	}
	_, err := c.Enrich(ctx, "Reply with JSON only: {\"summary\":\"ok\",\"cves\":[],\"techniques\":[],\"malware\":[],\"tools\":[],\"infrastructure\":[]}")
	return err
}
