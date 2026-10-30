package ingest

import (
	"encoding/json"

	"github.com/ami/ami/service/internal/pipeline"
)

// SourceConfig is the JSON blob on collect_source.config_json.
type SourceConfig struct {
	// direct_url
	ArticleURLs []string `json:"article_urls"`

	// list-based types
	ListURL     string         `json:"list_url"`
	EntryURL    string         `json:"entry_url"`
	ArticleBase string         `json:"article_base"`
	LinkExtract    LinkExtractCfg              `json:"link_extract"`
	Pagination     PaginationCfg               `json:"pagination"`
	TitleExtract   []pipeline.TitleExtractRule `json:"title_extract"`
	TitleBlocklist []string                    `json:"title_blocklist"`

	// Weekly WeChat digest: expand 情报来源 links from 攻击团伙/恶意代码 sections.
	DigestParser       string `json:"digest_parser"`        // wechat_weekly_threat
	ImportWeeklyDigest bool   `json:"import_weekly_digest"` // also import the digest page itself
	DigestSourceMax    int    `json:"digest_source_max"`    // cap child URLs expanded per run

	// legacy
	FeedURL        string `json:"feed_url"`
	LinkSelector   string `json:"link_selector"`
	MaxItemsPerRun int    `json:"max_items_per_run"`
	TimeoutSec     int    `json:"timeout_sec"`
	DelayMS        int    `json:"delay_ms"`
}

type LinkExtractCfg struct {
	MustContain      string   `json:"must_contain"`
	TitleMustContain string   `json:"title_must_contain"`
	ExcludeContains  []string `json:"exclude_contains"`
	PreferPathRegex  string   `json:"prefer_path_regex"`
	Selector         string   `json:"selector"`
}

type PaginationCfg struct {
	Type            string `json:"type"` // none | path | scroll | wechat_album
	PathTemplate    string `json:"path_template"`
	StartPage       int    `json:"start_page"`
	MaxPages        int    `json:"max_pages"`
	StopOnEmptyList bool   `json:"stop_on_empty_list"`
	ScrollSteps     int    `json:"scroll_steps"`
	ScrollWaitMS    int    `json:"scroll_wait_ms"`
	RenderMode      string `json:"render_mode"`
}

func ParseSourceConfig(raw string) SourceConfig {
	var cfg SourceConfig
	_ = json.Unmarshal([]byte(raw), &cfg)
	if cfg.MaxItemsPerRun <= 0 {
		cfg.MaxItemsPerRun = 10
	}
	if cfg.TimeoutSec <= 0 {
		cfg.TimeoutSec = 30
	}
	// legacy link_selector -> must_contain
	if cfg.LinkExtract.MustContain == "" && cfg.LinkSelector != "" {
		cfg.LinkExtract.MustContain = cfg.LinkSelector
	}
	return cfg
}

// TitleOpts builds pipeline title extraction options from this source config.
func (cfg SourceConfig) TitleOpts() *pipeline.TitleExtractOpts {
	if len(cfg.TitleExtract) == 0 && len(cfg.TitleBlocklist) == 0 {
		return nil
	}
	return &pipeline.TitleExtractOpts{
		Rules:     cfg.TitleExtract,
		Blocklist: cfg.TitleBlocklist,
	}
}
