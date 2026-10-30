package pipeline

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/ami/ami/service/internal/artifacts"
	"github.com/ami/ami/service/internal/llm"
	"github.com/ami/ami/service/internal/opslog"
	"github.com/ami/ami/service/internal/merge"
	"github.com/ami/ami/service/internal/store"
)

type Worker struct {
	DB    *store.DB
	Files *artifacts.Store
}

func (w *Worker) Run(ctx context.Context, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			w.tick()
		}
	}
}

// Tick runs one batch of pipeline jobs (exported for startup drain).
func (w *Worker) Tick() {
	w.tick()
}

func (w *Worker) tick() {
	for i := 0; i < 8; i++ {
		jobID, intelID, stage, err := w.DB.ClaimNextJob()
		if err != nil {
			log.Printf("pipeline claim: %v", err)
			return
		}
		if jobID == 0 {
			return
		}
		if stop, err := w.DB.IntelStopRequested(intelID); err == nil && stop {
			_ = w.DB.CompleteJob(jobID)
			_ = w.DB.ApplyIntelStop(intelID)
			continue
		}
		stageErr := w.runStage(jobID, intelID, stage)
		if stop, _ := w.DB.IntelStopRequested(intelID); stop {
			_ = w.DB.CompleteJob(jobID)
			_ = w.DB.ApplyIntelStop(intelID)
			continue
		}
		if stageErr != nil {
			log.Printf("pipeline intel=%d stage=%s: %v", intelID, stage, stageErr)
			opslog.Error("pipeline", stageErr.Error(), fmt.Sprintf("intelId=%d stage=%s", intelID, stage))
			_ = w.DB.FailJob(jobID, intelID, stageErr.Error())
		} else {
			_ = w.DB.CompleteJob(jobID)
		}
	}
}

func (w *Worker) runStage(jobID, intelID int64, stage string) error {
	switch stage {
	case "parse":
		return w.doParse(intelID)
	case "extract":
		return w.doExtract(intelID)
	default:
		return w.DB.FailJob(jobID, intelID, "unknown stage: "+stage)
	}
}

func (w *Worker) doParse(intelID int64) error {
	row, err := w.DB.GetIntel(intelID)
	if err != nil || row == nil {
		return err
	}
	_ = w.DB.SetIngestStatus(intelID, "parsing", row.KnowledgeAction, nil, "", "", nil)

	raw, err := w.Files.ReadRaw(intelID, 1)
	if err != nil {
		return err
	}
	html := string(raw)
	archived, assetN, _ := ArchivePageAssets(w.Files, intelID, row.SourceURL, raw)
	if len(archived) > 0 {
		html = string(archived)
		if _, _, _, err := w.Files.WriteRawHTML(intelID, 1, row.SourceURL, []byte(html)); err != nil {
			log.Printf("intel=%d rewrite archived html: %v", intelID, err)
		}
	}
	md := HTMLToMarkdown(html)
	if md == "" {
		md = "(empty)"
	}
	mdRel, err := w.Files.WriteParsedMarkdown(intelID, "v1", md)
	if err != nil {
		return err
	}
	snip := Snippet(md, 512)
	parseVer := "html-v2"
	if assetN > 0 {
		parseVer = fmt.Sprintf("html-v2+assets(%d)", assetN)
	}
	if err := w.DB.UpdateIntelAfterParse(intelID, snip, parseVer, mdRel); err != nil {
		return err
	}
	if disclosed := ExtractDisclosedAtFromHTML(html); disclosed != "" {
		_ = w.DB.UpdateIntelDisclosedAt(intelID, disclosed)
	}
	titleOpts := TitleOptsForIntel(w.DB, row.CollectSourceName, row.SourceURL)
	bl := titleOpts.blocklist()
	if htmlTitle := ExtractTitleFromHTML(html, titleOpts); isPlausibleWithBlocklist(htmlTitle, bl) {
		if htmlTitle != row.Title {
			_ = w.DB.UpdateIntelTitle(intelID, htmlTitle)
		}
	} else if ShouldRefreshIntelTitle(row.Title, row.SourceURL, htmlTitle, titleOpts) && htmlTitle != "" {
		_ = w.DB.UpdateIntelTitle(intelID, htmlTitle)
	} else if !isPlausibleWithBlocklist(row.Title, bl) {
		if t := TitleFromURL(row.SourceURL); t != "" {
			_ = w.DB.UpdateIntelTitle(intelID, t)
		}
	}
	_ = SyncFiguresFromHTML(w.Files, w.DB, intelID, row.SourceURL, []byte(html))
	return w.DB.EnqueueJob(intelID, "extract")
}

func (w *Worker) doExtract(intelID int64) error {
	row, err := w.DB.GetIntel(intelID)
	if err != nil || row == nil {
		return err
	}
	_ = w.DB.SetIngestStatus(intelID, "extracting", row.KnowledgeAction, nil, "", "", nil)

	bodyPath, _ := w.DB.ParsedLLMInputPath(intelID)
	var text string
	if bodyPath != "" {
		if b, err := w.Files.ReadParsedText(w.DB.DataDir(), bodyPath); err == nil {
			text = string(b)
		}
	}
	if text == "" && row.BodyPreview.Valid {
		text = row.BodyPreview.String
	}

	rawHTML, _ := w.Files.ReadRaw(intelID, 1)
	titleOpts := TitleOptsForIntel(w.DB, row.CollectSourceName, row.SourceURL)
	title := ResolveIntelTitle(row.Title, row.SourceURL, rawHTML, titleOpts)
	if title != row.Title {
		_ = w.DB.UpdateIntelTitle(intelID, title)
		row.Title = title
	}

	cves := ExtractCVEs(text)
	var enrichResult *llm.EnrichResult
	fallbackSummary := Snippet(text, 280)
	enrichVersion := "rules-v1"
	if out, err := llm.EnrichWithDB(w.DB, context.Background(), text); err == nil && out != nil && out.Result != nil {
		enrichVersion = out.VersionTag
		enrichResult = out.Result
		cves = mergeCVEs(cves, out.Result.SuggestedCves)
		opslog.Info("pipeline", "llm enrich ok", fmt.Sprintf("intelId=%d provider=%s", intelID, out.Provider))
	} else if err != nil {
		log.Printf("llm enrich intel=%d: %v", intelID, err)
		opslog.Warn("pipeline", "llm enrich failed (tried model list)", fmt.Sprintf("intelId=%d err=%v", intelID, err))
	}
	if enrichResult == nil {
		enrichResult = &llm.EnrichResult{
			Summary:             fallbackSummary,
			SuggestedCves:       cves,
			SuggestedTechniques: []string{},
			SuggestedMalware:    []string{},
		}
		enrichResult.Quality = llm.ValidateEnrich(enrichResult)
	}
	ai := aiExtractFromEnrich(enrichResult, fallbackSummary, cves)
	if len(cves) > 0 {
		ai.SuggestedTechniques = appendUnique(ai.SuggestedTechniques, "T1190")
	}

	existingKids, _ := w.DB.KnowledgeIDsForIntel(intelID)
	var kid int64
	action := "created"
	role := "primary"
	if len(existingKids) > 0 {
		kid = existingKids[0]
		action = "updated"
		role = "supporting"
	} else {
		prior, err := w.DB.PriorIntelForMerge(intelID)
		if err != nil {
			return err
		}
		var priorMerge []struct {
			Title        string
			KnowledgeIDs []int64
		}
		for _, p := range prior {
			priorMerge = append(priorMerge, struct {
				Title        string
				KnowledgeIDs []int64
			}{p.Title, p.KnowledgeIDs})
		}
		kid = merge.FindKnowledgeID(row.Title, priorMerge)
	}
	if kid > 0 {
		existing, _ := w.DB.GetKnowledge(kid)
		pl := store.KnowledgePayload{}
		if existing != nil {
			pl, _ = store.ParsePayload(existing.PayloadJSON)
		}
		pl = applyEnrichToPayload(pl, enrichResult, cves)
		pl.ThreatActors = mergeThreatActors(pl.ThreatActors, ExtractThreatActors(title+"\n"+text))
		kTitle := pickKnowledgeTitle(title, existing)
		_ = w.DB.UpdateKnowledgeSummary(kid, kTitle, ai.Summary, pl)
	} else {
		industries := []string{"互联网"}
		pl := applyEnrichToPayload(store.KnowledgePayload{Techniques: []map[string]any{}}, enrichResult, cves)
		pl.ThreatActors = mergeThreatActors(pl.ThreatActors, ExtractThreatActors(title+"\n"+text))
		var err error
		kid, err = w.DB.CreateKnowledge(row.Title, ai.Summary, "Enterprise", industries, pl)
		if err != nil {
			return err
		}
		role = "primary"
	}
	if err := w.DB.LinkIntel(kid, intelID, role); err != nil {
		return err
	}

	conf := 0.75
	if len(cves) > 0 {
		conf = 0.88
	}
	return w.DB.SetIngestStatus(intelID, "succeeded", action, &conf, "", enrichVersion, ai)
}

func pickKnowledgeTitle(intelTitle string, existing *store.KnowledgeRow) string {
	if merge.ShouldMergeByTitle(intelTitle) {
		if existing == nil || !merge.ShouldMergeByTitle(existing.Title) || existing.Title != intelTitle {
			return intelTitle
		}
	}
	if existing != nil && merge.ShouldMergeByTitle(existing.Title) {
		return existing.Title
	}
	if merge.ShouldMergeByTitle(intelTitle) {
		return intelTitle
	}
	return NormalizeTitle(intelTitle)
}

func mergeCVEs(a, b []string) []string {
	return appendUnique(a, b...)
}

func appendUnique(base []string, more ...string) []string {
	seen := make(map[string]struct{}, len(base))
	for _, s := range base {
		seen[s] = struct{}{}
	}
	for _, s := range more {
		s = strings.TrimSpace(s)
		if s == "" {
			continue
		}
		if _, ok := seen[s]; ok {
			continue
		}
		seen[s] = struct{}{}
		base = append(base, s)
	}
	return base
}

// ProcessImport writes raw artifact and enqueues parse job.
func ProcessImport(db *store.DB, files *artifacts.Store, intelID int64, sourceURL string, html []byte) error {
	rel, sha, size, err := files.WriteRawHTML(intelID, 1, sourceURL, html)
	if err != nil {
		return err
	}
	if err := db.InsertArtifact(intelID, 1, "html", rel, sha, size, "text/html"); err != nil {
		return err
	}
	preview := Snippet(PrepareArticleText(string(html)), 2048)
	if err := db.UpdateIntelAfterRaw(intelID, preview); err != nil {
		return err
	}
	return db.EnqueueJob(intelID, "parse")
}
