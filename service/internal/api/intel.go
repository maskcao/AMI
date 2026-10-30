package api

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/ami/ami/service/internal/pipeline"
	"github.com/ami/ami/service/internal/store"
)

func (s *Server) listIntel(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	p := store.IntelListParams{
		SourceName:   q.Get("sourceName"),
		IngestStatus: q.Get("ingestStatus"),
		SortOrder:    q.Get("sortOrder"),
		Limit:        200,
	}
	if kid := q.Get("knowledgeId"); kid != "" {
		if id, err := parseInt64(kid); err == nil {
			p.KnowledgeID = id
		}
	}
	rows, total, err := s.DB.ListIntel(p)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}
	intelIDs := make([]int64, len(rows))
	for i := range rows {
		intelIDs[i] = rows[i].ID
	}
	actorsByIntel, _ := s.DB.ThreatActorsByIntelIDs(intelIDs)
	items := make([]map[string]any, 0, len(rows))
	for _, row := range rows {
		items = append(items, intelSummary(s.DB, row, actorsByIntel[row.ID]))
	}
	writeJSON(w, http.StatusOK, map[string]any{"items": items, "total": total})
}

func (s *Server) getIntel(w http.ResponseWriter, r *http.Request) {
	id := pathID(r)
	row, err := s.DB.GetIntel(id)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}
	if row == nil {
		writeErr(w, http.StatusNotFound, "not found")
		return
	}
	detail, err := s.intelDetail(row, false)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, detail)
}

func (s *Server) getIntelOriginal(w http.ResponseWriter, r *http.Request) {
	id := pathID(r)
	row, err := s.DB.GetIntel(id)
	if err != nil || row == nil {
		writeErr(w, http.StatusNotFound, "not found")
		return
	}
	if row.HasRaw {
		b, err := s.Files.ReadRaw(id, 1)
		if err != nil {
			writeErr(w, http.StatusInternalServerError, err.Error())
			return
		}
		b = prepareSnapshotHTML(b, r, id)
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, _ = w.Write(b)
		return
	}
	if row.BodyPreview.Valid {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		_, _ = w.Write([]byte(row.BodyPreview.String))
		return
	}
	writeErr(w, http.StatusNotFound, "no original")
}

type importIntelReq struct {
	Title             string `json:"title"`
	SourceURL         string `json:"sourceUrl"`
	SourceName        string `json:"sourceName"`
	CollectSourceName string `json:"collectSourceName"`
	HTML              string `json:"html"`
}

func (s *Server) importIntel(w http.ResponseWriter, r *http.Request) {
	var req importIntelReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeErr(w, http.StatusBadRequest, "invalid json")
		return
	}
	req.Title = strings.TrimSpace(req.Title)
	if req.Title == "" || req.HTML == "" {
		writeErr(w, http.StatusBadRequest, "title and html required")
		return
	}
	if req.SourceName == "" {
		req.SourceName = "import"
	}
	if req.CollectSourceName == "" {
		req.CollectSourceName = "manual_upload"
	}
	id, err := s.DB.CreateIntel(req.Title, req.SourceURL, req.SourceName, req.CollectSourceName)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}
	if err := pipeline.ProcessImport(s.DB, s.Files, id, req.SourceURL, []byte(req.HTML)); err != nil {
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}
	row, _ := s.DB.GetIntel(id)
	detail, _ := s.intelDetail(row, false)
	writeJSON(w, http.StatusCreated, detail)
}

func (s *Server) searchIOC(w http.ResponseWriter, r *http.Request) {
	q := strings.TrimSpace(strings.ToLower(r.URL.Query().Get("q")))
	if q == "" {
		writeJSON(w, http.StatusOK, []any{})
		return
	}
	rows, err := s.DB.Query(`
		SELECT i.type, i.value, i.confidence, i.intel_id, ir.title
		FROM intel_ioc i JOIN intel_report ir ON ir.id = i.intel_id
		WHERE LOWER(i.value) LIKE ? OR LOWER(i.type) LIKE ? LIMIT 50`, "%"+q+"%", "%"+q+"%")
	if err != nil {
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}
	defer rows.Close()
	var hits []map[string]any
	for rows.Next() {
		var typ, val, title string
		var intelID int64
		var conf sql.NullString
		if err := rows.Scan(&typ, &val, &conf, &intelID, &title); err != nil {
			writeErr(w, http.StatusInternalServerError, err.Error())
			return
		}
		kids, _ := s.DB.KnowledgeIDsForIntel(intelID)
		hit := map[string]any{
			"ioc":           map[string]any{"type": typ, "value": val},
			"intelId":       intelID,
			"intelTitle":    title,
			"knowledgeIds":  kids,
		}
		if conf.Valid {
			hit["ioc"] = map[string]any{"type": typ, "value": val, "confidence": conf.String}
		}
		hits = append(hits, hit)
	}
	if hits == nil {
		hits = []map[string]any{}
	}
	writeJSON(w, http.StatusOK, hits)
}

func intelSummary(db *store.DB, row store.IntelRow, threatActors []string) map[string]any {
	kids, _ := db.KnowledgeIDsForIntel(row.ID)
	if kids == nil {
		kids = []int64{}
	}
	if threatActors == nil {
		threatActors = []string{}
	}
	iocN, _ := db.IOCCount(row.ID)
	m := map[string]any{
		"id":                row.ID,
		"title":             row.Title,
		"sourceUrl":         row.SourceURL,
		"disclosedAt":       nullStrJSON(row.DisclosedAt),
		"sourceName":        row.SourceName,
		"collectSourceName": row.CollectSourceName,
		"ingestStatus":      row.IngestStatus,
		"collectedAt":       row.CollectedAt,
		"knowledgeIds":      kids,
		"knowledgeAction":   row.KnowledgeAction,
		"extractConfidence": nullFloatJSON(row.ExtractConfidence),
		"iocCount":          iocN,
		"threatActors":      threatActors,
		"stopRequested":     row.StopRequested,
	}
	if row.PipelineError.Valid {
		m["pipelineError"] = row.PipelineError.String
	}
	return m
}

func (s *Server) intelDetail(row *store.IntelRow, includeFullHTML bool) (map[string]any, error) {
	actors, _ := s.DB.ThreatActorsByIntelIDs([]int64{row.ID})
	m := intelSummary(s.DB, *row, actors[row.ID])
	m["sourceUrl"] = row.SourceURL
	iocs, _ := s.DB.IOCsForIntel(row.ID)
	figures, _ := s.DB.FiguresForIntel(row.ID)
	m["iocs"] = iocs
	if iocs == nil {
		m["iocs"] = []store.IntelIOC{}
	}
	m["figures"] = figures
	if figures == nil {
		m["figures"] = []store.IntelFigure{}
	}
	parseV := "html-v1"
	if row.ParseVersion.Valid {
		parseV = row.ParseVersion.String
	}
	enrichV := ""
	if row.EnrichVersion.Valid {
		enrichV = row.EnrichVersion.String
	}
	m["parseVersion"] = parseV
	m["enrichVersion"] = enrichV
	m["hasRaw"] = row.HasRaw
	m["hasParsed"] = row.HasParsed
	var ai any = nil
	if row.AIExtractJSON.Valid && row.AIExtractJSON.String != "" {
		var ex store.AIExtract
		_ = json.Unmarshal([]byte(row.AIExtractJSON.String), &ex)
		ai = map[string]any{
			"summary":             ex.Summary,
			"suggestedTechniques": emptySlice(ex.SuggestedTechniques),
			"suggestedMalware":    emptySlice(ex.SuggestedMalware),
			"suggestedCves":       emptySlice(ex.SuggestedCves),
			"tools":               emptySlice(ex.Tools),
			"infrastructure":      emptySlice(ex.Infrastructure),
			"malwareClasses":      emptySlice(ex.MalwareClasses),
			"qualityHints":        emptySlice(ex.QualityHints),
			"qualityOk":           ex.QualityOK,
		}
	}
	m["aiExtract"] = ai
	html := ""
	if includeFullHTML && row.HasRaw {
		if b, err := s.Files.ReadRaw(row.ID, 1); err == nil {
			html = string(b)
		}
	} else if row.BodyPreview.Valid {
		html = row.BodyPreview.String
	}
	m["originalHtml"] = html
	return m, nil
}

func nullStrJSON(v sql.NullString) any {
	if !v.Valid {
		return nil
	}
	return v.String
}

func nullFloatJSON(v sql.NullFloat64) any {
	if !v.Valid {
		return nil
	}
	return v.Float64
}
