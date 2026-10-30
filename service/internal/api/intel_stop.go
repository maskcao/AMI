package api

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/ami/ami/service/internal/store"
)

func (s *Server) stopIntel(w http.ResponseWriter, r *http.Request) {
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
	outcome, err := s.DB.RequestIntelStop(id)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"id": id, "outcome": outcome})
}

type stopBatchRequest struct {
	IDs   []int64 `json:"ids"`
	Scope string  `json:"scope"`
}

// stopIntelBatch stops selected ids, or a scope: "queued" (default) or "active"
// (queued + parsing + extracting).
func (s *Server) stopIntelBatch(w http.ResponseWriter, r *http.Request) {
	var body stopBatchRequest
	if r.Body != nil {
		_ = json.NewDecoder(r.Body).Decode(&body)
	}
	ids := body.IDs
	if len(ids) == 0 {
		scope := strings.TrimSpace(body.Scope)
		if scope == "" {
			scope = "queued"
		}
		if scope != "queued" && scope != "active" {
			writeErr(w, http.StatusBadRequest, "scope must be queued or active")
			return
		}
		var err error
		ids, err = s.DB.IntelIDsForStop(scope)
		if err != nil {
			writeErr(w, http.StatusInternalServerError, err.Error())
			return
		}
	}

	stopped, stopping, skipped, failed := 0, 0, 0, 0
	for _, id := range ids {
		outcome, err := s.DB.RequestIntelStop(id)
		if err != nil {
			failed++
			continue
		}
		switch outcome {
		case store.IntelStopStopped:
			stopped++
		case store.IntelStopStopping:
			stopping++
		default:
			skipped++
		}
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"stopped": stopped, "stopping": stopping, "skipped": skipped, "failed": failed,
	})
}
