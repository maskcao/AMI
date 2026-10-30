package api

import (
	"net/http"
)

func (s *Server) dupScanKnowledge(w http.ResponseWriter, _ *http.Request) {
	updated, err := s.DB.BackfillDupCandidates()
	if err != nil {
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"flagged": updated})
}
