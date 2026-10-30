package api

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/ami/ami/service/internal/ingest"
	"github.com/ami/ami/service/internal/opslog"
	"github.com/ami/ami/service/internal/store"
	"github.com/go-chi/chi/v5"
)

func (s *Server) testCollectArticle(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	src, err := s.DB.GetCollectSource(id)
	if err != nil || src == nil {
		writeErr(w, http.StatusNotFound, "not found")
		return
	}
	var body struct {
		URL            string `json:"url"`
		WaitPipeline   bool   `json:"waitPipeline"`
	}
	if r.Body != nil {
		_ = json.NewDecoder(r.Body).Decode(&body)
	}
	url := strings.TrimSpace(body.URL)
	if url == "" {
		url = pilotDefaultArticleURL(src)
	}
	if url == "" {
		writeErr(w, http.StatusBadRequest, "url required (or configure article_urls / list_url)")
		return
	}
	wait := true
	if r.Body != nil {
		wait = body.WaitPipeline
	}
	opslog.Info("api", "pilot article test", "sourceId="+id+" url="+url)
	res, err := ingest.RunArticlePilot(s.DB, s.Files, src, url, wait)
	if err != nil {
		writeErr(w, http.StatusBadRequest, err.Error())
		return
	}
	src2, _ := s.DB.GetCollectSource(id)
	writeJSON(w, http.StatusOK, map[string]any{
		"passed":        res.Quality.Passed,
		"intelId":       res.IntelID,
		"sourceUrl":     res.SourceURL,
		"imported":      res.Imported,
		"reprocessed":   res.Reprocessed,
		"ingestStatus":  res.IngestStatus,
		"checks":        res.Quality.Checks,
		"collectSource": collectSourceJSON(*src2),
	})
}

func (s *Server) testCollectList(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	src, err := s.DB.GetCollectSource(id)
	if err != nil || src == nil {
		writeErr(w, http.StatusNotFound, "not found")
		return
	}
	if !ingest.TypeNeedsListPilot(src.Type) {
		_ = s.DB.MarkListPilotSkippedForDirectTypes(id, src.Type)
		src2, _ := s.DB.GetCollectSource(id)
		writeJSON(w, http.StatusOK, map[string]any{
			"passed":        true,
			"skipped":       true,
			"message":       "此采集类型无需分页试采",
			"collectSource": collectSourceJSON(*src2),
		})
		return
	}
	var body struct {
		MaxPages int `json:"maxPages"`
	}
	_ = json.NewDecoder(r.Body).Decode(&body)
	cfg := ingest.ParseSourceConfig(src.ConfigJSON)
	maxPages := body.MaxPages
	if maxPages <= 0 {
		maxPages = cfg.Pagination.MaxPages
		if maxPages <= 0 {
			maxPages = 2
		}
	}
	opslog.Info("api", "pilot list test", "sourceId="+id)
	disc, err := ingest.DiscoverArticleLinks(cfg, maxPages)
	if err != nil {
		writeErr(w, http.StatusBadRequest, err.Error())
		return
	}
	passed := disc.LinkCount > 0
	if passed {
		_ = s.DB.SetCollectSourcePilotListPassed(id)
	}
	src2, _ := s.DB.GetCollectSource(id)
	writeJSON(w, http.StatusOK, map[string]any{
		"passed":          passed,
		"pageUrlsScanned": disc.PageURLsScanned,
		"linkCount":       disc.LinkCount,
		"sampleUrls":      disc.SampleURLs,
		"collectSource":   collectSourceJSON(*src2),
	})
}

func (s *Server) unlockCollectBatch(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	src, err := s.DB.GetCollectSource(id)
	if err != nil || src == nil {
		writeErr(w, http.StatusNotFound, "not found")
		return
	}
	if ok, msg := s.DB.CanUnlockBatch(src); !ok {
		writeErr(w, http.StatusBadRequest, msg)
		return
	}
	if err := s.DB.SetCollectSourceBatchUnlocked(id, true); err != nil {
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}
	src2, _ := s.DB.GetCollectSource(id)
	opslog.Info("api", "collect batch unlocked", "sourceId="+id)
	writeJSON(w, http.StatusOK, map[string]any{
		"ok":            true,
		"collectSource": collectSourceJSON(*src2),
	})
}

func pilotDefaultArticleURL(src *store.CollectSourceFull) string {
	cfg := ingest.ParseSourceConfig(src.ConfigJSON)
	for _, u := range cfg.ArticleURLs {
		u = strings.TrimSpace(u)
		if u != "" {
			return u
		}
	}
	if u := strings.TrimSpace(cfg.ListURL); u != "" {
		return ""
	}
	return ""
}
