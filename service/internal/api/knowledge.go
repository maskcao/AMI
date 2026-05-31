package api

import (
	"encoding/json"
	"net/http"

	"github.com/ami/ami/service/internal/store"
)

func (s *Server) listKnowledge(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	p := store.KnowledgeListParams{
		Query: q.Get("q"), Status: q.Get("status"), Limit: 200,
		Matrix: q.Get("matrix"),
	}
	if ind := q.Get("industry"); ind != "" {
		p.Industry = ind
	}
	if mc := q.Get("malwareClass"); mc != "" {
		p.MalwareClass = mc
	}
	if q.Get("needsReview") == "true" {
		t := true
		p.NeedsReview = &t
	}
	rows, total, err := s.DB.ListKnowledge(p)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}
	items := make([]map[string]any, 0, len(rows))
	for _, row := range rows {
		items = append(items, knowledgeSummary(row))
	}
	writeJSON(w, http.StatusOK, map[string]any{"items": items, "total": total})
}

func (s *Server) getKnowledge(w http.ResponseWriter, r *http.Request) {
	id := pathID(r)
	row, err := s.DB.GetKnowledge(id)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}
	if row == nil {
		writeErr(w, http.StatusNotFound, "not found")
		return
	}
	writeJSON(w, http.StatusOK, s.buildKnowledgeDetail(*row))
}

func knowledgeSummary(row store.KnowledgeRow) map[string]any {
	var industries []string
	_ = json.Unmarshal([]byte(row.TargetIndustriesJSON), &industries)
	pl, _ := store.ParsePayload(row.PayloadJSON)
	dims := map[string]int{
		"techniques":      len(pl.Techniques),
		"tools":           len(pl.Tools),
		"infrastructure":  len(pl.Infrastructure),
		"vulnerabilities": len(pl.Vulnerabilities),
		"malware":         len(pl.Malware),
	}
	return map[string]any{
		"id":               row.ID,
		"title":            row.Title,
		"summary":          nullStrJSON(row.Summary),
		"status":           row.Status,
		"needsReview":      row.NeedsReview,
		"targetIndustries": industries,
		"attackMatrix":     row.AttackMatrix,
		"malware":          normalizeMalwarePayload(emptySlice(pl.Malware)),
		"threatActors":     emptyThreatActors(pl.ThreatActors),
		"qualityScore":     row.QualityScore,
		"intelCount":       row.IntelCount,
		"updatedAt":        row.UpdatedAt,
		"dimensions":       dims,
	}
}

func emptyThreatActors(s []string) []string {
	if s == nil {
		return []string{}
	}
	return s
}
