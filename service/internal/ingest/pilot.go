package ingest

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/ami/ami/service/internal/artifacts"
	"github.com/ami/ami/service/internal/pipeline"
	"github.com/ami/ami/service/internal/store"
)

// ListDiscoverResult is the outcome of a list/pagination pilot (no import).
type ListDiscoverResult struct {
	PageURLsScanned int      `json:"pageUrlsScanned"`
	LinkCount       int      `json:"linkCount"`
	SampleURLs      []string `json:"sampleUrls"`
}

// ArticlePilotResult is the outcome of a single-article pilot import + optional pipeline wait.
type ArticlePilotResult struct {
	IntelID       int64                    `json:"intelId"`
	SourceURL     string                   `json:"sourceUrl"`
	Imported      bool                     `json:"imported"`
	Reprocessed   bool                     `json:"reprocessed"`
	IngestStatus  string                   `json:"ingestStatus"`
	Quality       pipeline.QualityReport   `json:"quality"`
}

// TypeNeedsListPilot reports whether list/pagination pilot is required before batch unlock.
func TypeNeedsListPilot(typ string) bool {
	switch typ {
	case "paginated_list", "threat_portal", "scheduled_url", "wechat_feed":
		return true
	default:
		return false
	}
}

// DiscoverArticleLinks fetches list pages and returns article URLs without importing.
func DiscoverArticleLinks(cfg SourceConfig, maxPages int) (*ListDiscoverResult, error) {
	if cfg.Pagination.Type == "scroll" {
		return nil, ScrollPaginationError()
	}
	if cfg.Pagination.Type == "wechat_album" {
		return discoverWeChatAlbumLinks(cfg, maxPages)
	}
	if cfg.Pagination.Type == "wordpress_ajax" {
		return discoverWordPressAjaxLinks(cfg, maxPages)
	}
	if maxPages > 0 {
		cfg.Pagination.MaxPages = maxPages
	}
	pageURLs := ExpandListPageURLs(cfg)
	if len(pageURLs) == 0 {
		return nil, fmt.Errorf("no list page URLs configured")
	}
	base := cfg.ArticleBase
	if base == "" && len(pageURLs) > 0 {
		base = pageURLs[0]
	}
	seen := map[string]struct{}{}
	var queue []listLink
	pagesScanned := 0
	for _, pageURL := range pageURLs {
		body, err := httpGet(pageURL, cfg.TimeoutSec)
		if err != nil {
			continue
		}
		pagesScanned++
		links, err := extractArticleLinks(body, base, cfg)
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

// RunArticlePilot imports one article URL and optionally waits for the pipeline + quality gate.
func RunArticlePilot(db *store.DB, files *artifacts.Store, src *store.CollectSourceFull, articleURL string, waitPipeline bool) (*ArticlePilotResult, error) {
	if src == nil {
		return nil, fmt.Errorf("collect source not found")
	}
	articleURL = strings.TrimSpace(articleURL)
	if articleURL == "" {
		return nil, fmt.Errorf("article url required")
	}
	cfg := ParseSourceConfig(src.ConfigJSON)
	html, err := fetchArticleHTML(articleURL, cfg.TimeoutSec)
	if err != nil {
		return nil, err
	}
	sourceName := labelForSource(src)
	title := pipeline.PickIngestTitle(html, articleURL, "", cfg.TitleOpts())
	if title == "" {
		title = articleURL
	}

	out := &ArticlePilotResult{SourceURL: articleURL}
	existingID, err := db.GetIntelIDBySourceURL(articleURL)
	if err != nil {
		return nil, err
	}
	var intelID int64
	if existingID > 0 {
		intelID = existingID
		out.Reprocessed = true
		if err := pipeline.ProcessImport(db, files, intelID, articleURL, html); err != nil {
			return nil, err
		}
	} else {
		id, err := db.CreateIntel(title, articleURL, sourceName, src.Name)
		if err != nil {
			return nil, err
		}
		intelID = id
		out.Imported = true
		if err := pipeline.ProcessImport(db, files, intelID, articleURL, html); err != nil {
			return nil, err
		}
	}
	out.IntelID = intelID

	if waitPipeline {
		st, err := waitIntelPipeline(db, files, intelID, 800)
		if err != nil {
			return out, err
		}
		out.IngestStatus = st
	} else {
		row, _ := db.GetIntel(intelID)
		if row != nil {
			out.IngestStatus = row.IngestStatus
		}
	}

	raw, _ := files.ReadRaw(intelID, 1)
	out.Quality = pipeline.RunIntelQualityGate(files, db, intelID, raw)
	if out.Quality.Passed {
		_ = db.SetCollectSourcePilotArticlePassed(src.ID)
		_ = db.MarkListPilotSkippedForDirectTypes(src.ID, src.Type)
	}
	return out, nil
}

func waitIntelPipeline(db *store.DB, files *artifacts.Store, intelID int64, maxTicks int) (string, error) {
	worker := &pipeline.Worker{DB: db, Files: files}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Minute)
	defer cancel()
	done := 0
	for tick := 0; tick < maxTicks; tick++ {
		select {
		case <-ctx.Done():
			return "", ctx.Err()
		default:
		}
		row, err := db.GetIntel(intelID)
		if err != nil {
			return "", err
		}
		if row == nil {
			return "", fmt.Errorf("intel %d not found", intelID)
		}
		if row.IngestStatus == "succeeded" || row.IngestStatus == "failed" {
			return row.IngestStatus, nil
		}
		worker.Tick()
		afterRow, _ := db.GetIntel(intelID)
		if afterRow != nil && (afterRow.IngestStatus == "succeeded" || afterRow.IngestStatus == "failed") {
			return afterRow.IngestStatus, nil
		}
		stillPending := row.IngestStatus == "queued" || row.IngestStatus == "parsing" || row.IngestStatus == "extracting"
		if !stillPending && row.IngestStatus != "succeeded" && row.IngestStatus != "failed" {
			return row.IngestStatus, nil
		}
		if stillPending {
			done = 0
		} else {
			done++
			if done >= 3 {
				return row.IngestStatus, nil
			}
		}
		time.Sleep(25 * time.Millisecond)
	}
	return "", fmt.Errorf("pipeline timeout for intel %d", intelID)
}

// RunCollectSourceWithGate runs full collect only when batch_unlocked and pilot checks pass.
func RunCollectSourceWithGate(db *store.DB, files *artifacts.Store, sourceID string) (*RunResult, error) {
	src, err := db.GetCollectSource(sourceID)
	if err != nil {
		return nil, err
	}
	if src == nil {
		return nil, fmt.Errorf("collect source not found")
	}
	if ok, msg := db.PilotReadyForBatch(src); !ok {
		return nil, fmt.Errorf("%s", msg)
	}
	return RunCollectSource(db, files, sourceID)
}
