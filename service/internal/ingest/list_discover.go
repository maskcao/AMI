package ingest

import (
	"log"
	"strings"
	"time"

	"github.com/ami/ami/service/internal/artifacts"
	"github.com/ami/ami/service/internal/pipeline"
	"github.com/ami/ami/service/internal/store"
)

func runDirectURL(db *store.DB, files *artifacts.Store, src *store.CollectSourceFull, cfg SourceConfig, res *RunResult) error {
	if len(cfg.ArticleURLs) == 0 {
		return nil
	}
	sourceName := labelForSource(src)
	for _, u := range cfg.ArticleURLs {
		u = strings.TrimSpace(u)
		if u == "" {
			continue
		}
		html, err := fetchArticleHTML(u, cfg.TimeoutSec)
		if err != nil {
			res.Failed++
			continue
		}
		title := pipeline.PickIngestTitle(html, u, "")
		if err := importOne(db, files, sourceName, src.Name, title, u, html, res); err != nil {
			log.Printf("ingest direct %s: %v", u, err)
			res.Failed++
		}
		sleepDelay(cfg.DelayMS)
	}
	return nil
}

func runListDiscover(db *store.DB, files *artifacts.Store, src *store.CollectSourceFull, cfg SourceConfig, res *RunResult) error {
	if cfg.Pagination.Type == "scroll" {
		return ScrollPaginationError()
	}
	pageURLs := ExpandListPageURLs(cfg)
	if len(pageURLs) == 0 {
		return nil
	}
	base := cfg.ArticleBase
	if base == "" && len(pageURLs) > 0 {
		base = pageURLs[0]
	}
	sourceName := labelForSource(src)
	seenArticles := map[string]struct{}{}
	var queue []listLink

	for _, pageURL := range pageURLs {
		body, err := httpGet(pageURL, cfg.TimeoutSec)
		if err != nil {
			log.Printf("ingest list page %s: %v", pageURL, err)
			res.Failed++
			continue
		}
		links, err := extractArticleLinks(body, base, cfg)
		if err != nil {
			return err
		}
		if cfg.Pagination.StopOnEmptyList && len(links) == 0 {
			break
		}
		for _, lk := range links {
			if _, ok := seenArticles[lk.URL]; ok {
				continue
			}
			seenArticles[lk.URL] = struct{}{}
			queue = append(queue, lk)
		}
		sleepDelay(cfg.DelayMS)
	}

	limit := cfg.MaxItemsPerRun
	if limit <= 0 {
		limit = 10
	}
	if len(queue) > limit {
		queue = queue[:limit]
	}

	for _, lk := range queue {
		html, err := fetchArticleHTML(lk.URL, cfg.TimeoutSec)
		if err != nil {
			res.Failed++
			continue
		}
		title := pipeline.PickIngestTitle(html, lk.URL, lk.Title)
		if err := importOne(db, files, sourceName, src.Name, title, lk.URL, html, res); err != nil {
			res.Failed++
		}
		sleepDelay(cfg.DelayMS)
	}
	return nil
}

func labelForSource(src *store.CollectSourceFull) string {
	if src.SourceLabel != "" {
		return src.SourceLabel
	}
	return src.Name
}

func sleepDelay(ms int) {
	if ms <= 0 {
		return
	}
	time.Sleep(time.Duration(ms) * time.Millisecond)
}

