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
		title := pipeline.PickIngestTitle(html, u, "", cfg.TitleOpts())
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
	if cfg.Pagination.Type == "wechat_album" {
		return runWeChatAlbumList(db, files, src, cfg, res)
	}
	if cfg.Pagination.Type == "wordpress_ajax" {
		return runWordPressAjaxList(db, files, src, cfg, res)
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

	importListQueue(db, files, sourceName, src.Name, cfg, queue, res)
	return nil
}

func runWeChatAlbumList(db *store.DB, files *artifacts.Store, src *store.CollectSourceFull, cfg SourceConfig, res *RunResult) error {
	queue, _, err := collectWeChatAlbumLinks(cfg, 0)
	if err != nil {
		return err
	}
	if len(queue) == 0 {
		return nil
	}
	queue, err = ExpandWeeklyDigestSourceLinks(cfg, queue)
	if err != nil {
		return err
	}
	sourceName := labelForSource(src)
	importListQueue(db, files, sourceName, src.Name, cfg, queue, res)
	return nil
}

func runWordPressAjaxList(db *store.DB, files *artifacts.Store, src *store.CollectSourceFull, cfg SourceConfig, res *RunResult) error {
	queue, _, err := collectWordPressAjaxLinks(cfg, 0)
	if err != nil {
		return err
	}
	if len(queue) == 0 {
		return nil
	}
	sourceName := labelForSource(src)
	importListQueue(db, files, sourceName, src.Name, cfg, queue, res)
	return nil
}

// importListQueue walks the full paginated queue. Skipped (already ingested) URLs do not
// count toward max_items_per_run so a later run can import links from page 2+.
func importListQueue(db *store.DB, files *artifacts.Store, sourceName, collectName string, cfg SourceConfig, queue []listLink, res *RunResult) {
	limit := cfg.MaxItemsPerRun
	if limit <= 0 {
		limit = 10
	}
	importedNew := 0
	for _, lk := range queue {
		if importedNew >= limit {
			break
		}
		exists, err := db.IntelExistsBySourceURL(lk.URL)
		if err != nil {
			res.Failed++
			continue
		}
		if exists {
			res.Skipped++
			continue
		}
		html, err := fetchArticleHTML(lk.URL, cfg.TimeoutSec)
		if err != nil {
			res.Failed++
			continue
		}
		title := pipeline.PickIngestTitle(html, lk.URL, lk.Title, cfg.TitleOpts())
		before := res.Imported
		if err := importOne(db, files, sourceName, collectName, title, lk.URL, html, res); err != nil {
			res.Failed++
			continue
		}
		if res.Imported > before {
			importedNew++
		}
		sleepDelay(cfg.DelayMS)
	}
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

