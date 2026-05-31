package api

import (
	"encoding/json"
	"net/http"

	"github.com/ami/ami/service/internal/corpus"
	"github.com/ami/ami/service/internal/store"
)

type knowledgePatchReq struct {
	Title            *string  `json:"title"`
	Summary          *string  `json:"summary"`
	TargetIndustries []string `json:"targetIndustries"`
}

func (s *Server) patchKnowledge(w http.ResponseWriter, r *http.Request) {
	id := pathID(r)
	var req knowledgePatchReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeErr(w, http.StatusBadRequest, "invalid json")
		return
	}
	patch := store.KnowledgePatch{
		Title:            req.Title,
		Summary:          req.Summary,
		TargetIndustries: req.TargetIndustries,
	}
	if err := s.DB.PatchKnowledge(id, patch); err != nil {
		writeErr(w, http.StatusNotFound, err.Error())
		return
	}
	s.respondKnowledge(w, id)
}

func (s *Server) publishKnowledge(w http.ResponseWriter, r *http.Request) {
	id := pathID(r)
	if err := s.DB.PublishKnowledge(id); err != nil {
		writeErr(w, http.StatusNotFound, err.Error())
		return
	}
	if _, err := corpus.ExportOnPublish(s.DB, id); err != nil {
		// non-fatal: publish still succeeds
		_ = err
	}
	s.respondKnowledge(w, id)
}

type mergeReq struct {
	SourceID int64 `json:"sourceId"`
	TargetID int64 `json:"targetId"`
}

func (s *Server) mergeKnowledge(w http.ResponseWriter, r *http.Request) {
	var req mergeReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeErr(w, http.StatusBadRequest, "invalid json")
		return
	}
	if req.SourceID == 0 || req.TargetID == 0 || req.SourceID == req.TargetID {
		writeErr(w, http.StatusBadRequest, "invalid merge ids")
		return
	}
	if err := s.DB.MergeKnowledge(req.SourceID, req.TargetID); err != nil {
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]bool{"ok": true})
}

func (s *Server) respondKnowledge(w http.ResponseWriter, id int64) {
	row, err := s.DB.GetKnowledge(id)
	if err != nil || row == nil {
		writeErr(w, http.StatusNotFound, "not found")
		return
	}
	writeJSON(w, http.StatusOK, s.buildKnowledgeDetail(*row))
}

func (s *Server) buildKnowledgeDetail(row store.KnowledgeRow) map[string]any {
	m := knowledgeSummary(row)
	pl, _ := store.ParsePayload(row.PayloadJSON)
	m["techniques"] = emptySlice(pl.Techniques)
	m["tools"] = emptySlice(pl.Tools)
	m["infrastructure"] = emptySlice(pl.Infrastructure)
	m["vulnerabilities"] = emptySlice(pl.Vulnerabilities)
	m["procedures"] = emptySlice(pl.Procedures)
	m["malware"] = normalizeMalwarePayload(emptySlice(pl.Malware))
	if pl.ThreatActors == nil {
		m["threatActors"] = []string{}
	} else {
		m["threatActors"] = pl.ThreatActors
	}
	m["qualityHints"] = emptySlice(pl.QualityHints)

	links, _ := s.DB.LinkedIntelForKnowledge(row.ID)
	linked := make([]map[string]any, 0, len(links))
	for _, l := range links {
		item := map[string]any{
			"id":                l.ID,
			"title":             l.Title,
			"sourceName":        l.SourceName,
			"sourceUrl":         l.SourceURL,
			"collectSourceName": l.CollectSourceName,
			"role":              l.Role,
			"hasRaw":            l.HasRaw,
			"hasParsed":         l.HasParsed,
		}
		if l.DisclosedAt.Valid {
			item["disclosedAt"] = l.DisclosedAt.String
		} else {
			item["disclosedAt"] = nil
		}
		linked = append(linked, item)
	}
	m["linkedIntel"] = linked

	figures, _ := s.DB.FiguresForKnowledge(row.ID)
	m["evidenceFigures"] = figures
	if figures == nil {
		m["evidenceFigures"] = []store.IntelFigure{}
	}

	if ir, _ := s.DB.LatestPipelineIntel(row.ID); ir != nil {
		conf := nullFloatJSON(ir.ExtractConfidence)
		m["pipeline"] = map[string]any{
			"lastIntelId":       ir.ID,
			"lastIntelTitle":    ir.Title,
			"knowledgeAction":   ir.KnowledgeAction,
			"extractConfidence": conf,
			"pipelinedAt":       ir.UpdatedAt,
		}
	} else {
		m["pipeline"] = pl.Pipeline
	}
	return m
}

func emptySlice[T any](s []T) []T {
	if s == nil {
		return []T{}
	}
	return s
}
