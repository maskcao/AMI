package ingest

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/ami/ami/service/internal/netutil"
)

var blogLoadmoreParamsRE = regexp.MustCompile(`blog_loadmore_params\s*=\s*(\{.*?\});`)

type wpLoadmoreBootstrap struct {
	AjaxURL        string
	PostsQuery     string
	Nonce          string
	MaxPage        int
	TrackingPrefix string
	Language       string
}

type wpLoadmoreJSON struct {
	HTML string `json:"html"`
}

// discoverWordPressAjaxLinks fetches article URLs via WordPress admin-ajax loadmore (Unit42-style tag pages).
func discoverWordPressAjaxLinks(cfg SourceConfig, maxPages int) (*ListDiscoverResult, error) {
	seed := strings.TrimSpace(cfg.ListURL)
	if seed == "" {
		seed = strings.TrimSpace(cfg.EntryURL)
	}
	if seed == "" {
		return nil, fmt.Errorf("list_url required for wordpress_ajax pagination")
	}
	body, err := httpGet(seed, cfg.TimeoutSec)
	if err != nil {
		return nil, fmt.Errorf("fetch list page: %w", err)
	}
	boot, err := parseWPLoadmoreBootstrap(body, seed)
	if err != nil {
		return nil, err
	}
	if maxPages <= 0 {
		maxPages = cfg.Pagination.MaxPages
	}
	if maxPages <= 0 {
		maxPages = boot.MaxPage
	}
	if maxPages <= 0 {
		maxPages = 5
	}
	base := strings.TrimSpace(cfg.ArticleBase)
	if base == "" {
		base = seed
	}

	seen := map[string]struct{}{}
	var queue []listLink
	pagesScanned := 0

	for page := 1; page <= maxPages; page++ {
		loadStatus := "init"
		if page > 1 {
			loadStatus = "more"
		}
		fragment, err := wpLoadmorePOST(boot, page, loadStatus, cfg.TimeoutSec)
		if err != nil {
			if page == 1 {
				return nil, err
			}
			break
		}
		pagesScanned++
		links, err := extractArticleLinks(fragment, base, cfg)
		if err != nil {
			return nil, err
		}
		if cfg.Pagination.StopOnEmptyList && len(links) == 0 {
			break
		}
		for _, lk := range links {
			if _, ok := seen[lk.URL]; ok {
				continue
			}
			seen[lk.URL] = struct{}{}
			queue = append(queue, lk)
		}
		sleepDelay(cfg.DelayMS)
	}

	sample := make([]string, 0, 10)
	for i, lk := range queue {
		if i >= 10 {
			break
		}
		sample = append(sample, lk.URL)
	}
	return &ListDiscoverResult{
		PageURLsScanned: pagesScanned,
		LinkCount:       len(queue),
		SampleURLs:      sample,
	}, nil
}

func parseWPLoadmoreBootstrap(html []byte, listURL string) (*wpLoadmoreBootstrap, error) {
	m := blogLoadmoreParamsRE.FindSubmatch(html)
	if len(m) < 2 {
		return nil, fmt.Errorf("blog_loadmore_params not found on list page (site may not use WordPress AJAX pagination)")
	}
	var raw struct {
		AjaxURL     string `json:"ajaxurl"`
		Posts       string `json:"posts"`
		Nonce       string `json:"nonce"`
		MaxPage     string `json:"max_page"`
		CurrentPage string `json:"current_page"`
	}
	if err := json.Unmarshal(m[1], &raw); err != nil {
		return nil, fmt.Errorf("parse blog_loadmore_params: %w", err)
	}
	if raw.AjaxURL == "" || raw.Posts == "" || raw.Nonce == "" {
		return nil, fmt.Errorf("incomplete blog_loadmore_params on list page")
	}
	maxPage, _ := strconv.Atoi(strings.TrimSpace(raw.MaxPage))
	tracking := extractHTMLAttr(string(html), `data-load-posts[^>]*data-tracking="([^"]+)"`)
	if tracking == "" {
		tracking = "post_tag"
	}
	lang := extractHTMLAttr(string(html), `data-load-posts[^>]*data-lang="([^"]+)"`)
	if lang == "" {
		lang = "en"
	}
	return &wpLoadmoreBootstrap{
		AjaxURL:        raw.AjaxURL,
		PostsQuery:     raw.Posts,
		Nonce:          raw.Nonce,
		MaxPage:        maxPage,
		TrackingPrefix: tracking,
		Language:       lang,
	}, nil
}

func extractHTMLAttr(html, rePat string) string {
	re := regexp.MustCompile(rePat)
	m := re.FindStringSubmatch(html)
	if len(m) >= 2 {
		return m[1]
	}
	return ""
}

func wpLoadmorePOST(boot *wpLoadmoreBootstrap, page int, loadStatus string, timeoutSec int) ([]byte, error) {
	if boot == nil {
		return nil, fmt.Errorf("missing wordpress ajax bootstrap")
	}
	form := url.Values{}
	form.Set("action", "loadmore")
	form.Set("query", boot.PostsQuery)
	form.Set("page", strconv.Itoa(page))
	form.Set("loadstatus", loadStatus)
	form.Set("nonce", boot.Nonce)
	form.Set("tracking_prefix", boot.TrackingPrefix)
	form.Set("language", boot.Language)
	form.Set("sort_by", "")

	if timeoutSec <= 0 {
		timeoutSec = 30
	}
	req, err := http.NewRequest(http.MethodPost, boot.AjaxURL, strings.NewReader(form.Encode()))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("User-Agent", netutil.DefaultUserAgent)
	client := &http.Client{Timeout: time.Duration(timeoutSec) * time.Second}
	res, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	if res.StatusCode >= 400 {
		return nil, fmt.Errorf("POST %s: %d", boot.AjaxURL, res.StatusCode)
	}
	const maxBody = 4 << 20
	data, err := io.ReadAll(io.LimitReader(res.Body, maxBody))
	if err != nil {
		return nil, err
	}
	var payload wpLoadmoreJSON
	if err := json.Unmarshal(data, &payload); err != nil {
		return nil, fmt.Errorf("wordpress ajax response: %w", err)
	}
	if strings.TrimSpace(payload.HTML) == "" {
		return nil, fmt.Errorf("wordpress ajax returned empty html for page %d", page)
	}
	return []byte(payload.HTML), nil
}

// collectWordPressAjaxLinks is used by full collect runs.
func collectWordPressAjaxLinks(cfg SourceConfig, maxPages int) ([]listLink, int, error) {
	disc, err := discoverWordPressAjaxLinks(cfg, maxPages)
	if err != nil {
		return nil, 0, err
	}
	if disc == nil || disc.LinkCount == 0 {
		return nil, disc.PageURLsScanned, nil
	}
	seed := strings.TrimSpace(cfg.ListURL)
	if seed == "" {
		seed = cfg.EntryURL
	}
	base := strings.TrimSpace(cfg.ArticleBase)
	if base == "" {
		base = seed
	}
	// Re-walk pages to preserve titles (discover only kept sample URLs).
	body, err := httpGet(seed, cfg.TimeoutSec)
	if err != nil {
		return nil, 0, err
	}
	boot, err := parseWPLoadmoreBootstrap(body, seed)
	if err != nil {
		return nil, 0, err
	}
	if maxPages <= 0 {
		maxPages = cfg.Pagination.MaxPages
	}
	if maxPages <= 0 {
		maxPages = boot.MaxPage
	}
	if maxPages <= 0 {
		maxPages = 5
	}
	seen := map[string]struct{}{}
	var out []listLink
	pages := 0
	for page := 1; page <= maxPages; page++ {
		loadStatus := "init"
		if page > 1 {
			loadStatus = "more"
		}
		fragment, err := wpLoadmorePOST(boot, page, loadStatus, cfg.TimeoutSec)
		if err != nil {
			if page == 1 {
				return nil, 0, err
			}
			break
		}
		pages++
		links, err := extractArticleLinks(fragment, base, cfg)
		if err != nil {
			return nil, pages, err
		}
		if cfg.Pagination.StopOnEmptyList && len(links) == 0 {
			break
		}
		for _, lk := range links {
			if _, ok := seen[lk.URL]; ok {
				continue
			}
			seen[lk.URL] = struct{}{}
			out = append(out, lk)
		}
		sleepDelay(cfg.DelayMS)
	}
	return out, pages, nil
}
