package api

import (
	"net/http"
)

func (s *Server) retryIntel(w http.ResponseWriter, r *http.Request) {
	id := pathID(r)
	row, err := s.DB.GetIntel(id)
	if err != nil || row == nil {
		writeErr(w, http.StatusNotFound, "not found")
		return
	}
	_ = s.DB.ClearPendingPipelineJobs(id)
	_ = s.DB.SetIngestStatus(id, "queued", "pending", nil, "", "", nil)
	if err := s.DB.EnqueueJob(id, "parse"); err != nil {
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}
	s.getIntel(w, r)
}
