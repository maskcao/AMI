package ingest

import (
	"fmt"
	"log"
	"strings"

	"github.com/ami/ami/service/internal/artifacts"
	"github.com/ami/ami/service/internal/opslog"
	"github.com/ami/ami/service/internal/pipeline"
	"github.com/ami/ami/service/internal/store"
)

type RunResult struct {
	Imported int `json:"imported"`
	Skipped  int `json:"skipped"`
	Failed   int `json:"failed"`
}

// RunCollectSource runs a full collect for the source (caller should enforce pilot gate when needed).
func RunCollectSource(db *store.DB, files *artifacts.Store, sourceID string) (*RunResult, error) {
	src, err := db.GetCollectSource(sourceID)
	if err != nil {
		return nil, err
	}
	if src == nil {
		return nil, fmt.Errorf("source not found")
	}
	if src.Type == "manual_upload" {
		return nil, fmt.Errorf("manual_upload cannot be scheduled")
	}

	cfg := ParseSourceConfig(src.ConfigJSON)
	opslog.Info("ingest", fmt.Sprintf("collect start: %s", src.Name), fmt.Sprintf("id=%s type=%s", sourceID, src.Type))

	var res RunResult
	var runErr error
	switch src.Type {
	case "direct_url":
		runErr = runDirectURL(db, files, src, cfg, &res)
	case "paginated_list", "threat_portal":
		runErr = runListDiscover(db, files, src, cfg, &res)
	case "rss":
		runErr = runRSS(db, files, src, cfg, &res)
	case "scheduled_url", "wechat_feed":
		runErr = runListDiscover(db, files, src, cfg, &res)
	default:
		runErr = fmt.Errorf("unsupported collect type: %s", src.Type)
	}

	status := "success"
	errMsg := ""
	if runErr != nil {
		status = "failed"
		errMsg = runErr.Error()
	} else if res.Failed > 0 && res.Imported > 0 {
		status = "partial"
	} else if res.Failed > 0 && res.Imported == 0 {
		status = "failed"
		errMsg = fmt.Sprintf("%d items failed", res.Failed)
	}
	_ = db.SetCollectSourceRunResult(sourceID, status, errMsg)
	if runErr != nil {
		opslog.Error("ingest", runErr.Error(), fmt.Sprintf("source=%s imported=%d failed=%d", sourceID, res.Imported, res.Failed))
		return &res, runErr
	}
	opslog.Info("ingest", fmt.Sprintf("collect done: %s", src.Name),
		fmt.Sprintf("imported=%d skipped=%d failed=%d status=%s", res.Imported, res.Skipped, res.Failed, status))
	return &res, nil
}

func runRSS(db *store.DB, files *artifacts.Store, src *store.CollectSourceFull, cfg SourceConfig, res *RunResult) error {
	if cfg.FeedURL == "" {
		return fmt.Errorf("config.feed_url required")
	}
	body, err := httpGet(cfg.FeedURL, cfg.TimeoutSec)
	if err != nil {
		return err
	}
	entries, err := parseFeedXML(body)
	if err != nil {
		return err
	}
	sourceName := labelForSource(src)
	for i, e := range entries {
		if i >= cfg.MaxItemsPerRun {
			break
		}
		if err := importOne(db, files, sourceName, src.Name, e.Title, e.Link, nil, res); err != nil {
			log.Printf("ingest rss item %s: %v", e.Link, err)
			res.Failed++
		}
	}
	return nil
}

func importOne(db *store.DB, files *artifacts.Store, sourceName, collectName, title, link string, html []byte, res *RunResult) error {
	if link != "" {
		exists, err := db.IntelExistsBySourceURL(link)
		if err != nil {
			return err
		}
		if exists {
			res.Skipped++
			return nil
		}
	}
	if len(html) == 0 {
		var err error
		html, err = httpGet(link, 30)
		if err != nil {
			return err
		}
	}
	if len(html) == 0 {
		return fmt.Errorf("empty body")
	}
	title = strings.TrimSpace(title)
	if title == "" {
		title = link
	}
	id, err := db.CreateIntel(title, link, sourceName, collectName)
	if err != nil {
		return err
	}
	if err := pipeline.ProcessImport(db, files, id, link, html); err != nil {
		return err
	}
	res.Imported++
	return nil
}
