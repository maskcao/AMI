package api

import (
	"net/http"
	"strconv"
)

func (s *Server) retryIntel(w http.ResponseWriter, r *http.Request) {
	id := pathID(r)
	row, err := s.DB.GetIntel(id)
	if err != nil || row == nil {
		writeErr(w, http.StatusNotFound, "not found")
		return
	}
	_ = s.DB.ClearPendingPipelineJobs(id)
	_ = s.DB.ClearIntelStop(id)
	_ = s.DB.SetIngestStatus(id, "queued", "pending", nil, "", "", nil)
	if err := s.DB.EnqueueJob(id, "parse"); err != nil {
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}
	s.getIntel(w, r)
}

// retryLLMFailed requeues the extract stage for intel whose enrichment fell
// back to rules (or was never LLM-enriched), capped per call because each item
// costs one LLM request.
func (s *Server) retryLLMFailed(w http.ResponseWriter, r *http.Request) {
	limit := 200
	if v := r.URL.Query().Get("limit"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 && n <= 500 {
			limit = n
		}
	}
	ids, remaining, err := s.DB.LLMFailedIntel(limit)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}
	if len(ids) == 0 {
		writeJSON(w, http.StatusOK, map[string]any{"requeued": 0, "remaining": remaining})
		return
	}
	if err := s.DB.RequeueLLMFailed(ids); err != nil {
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"requeued": len(ids), "remaining": remaining - len(ids)})
}
