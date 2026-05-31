package api

import "net/http"

func (s *Server) purgeAllData(w http.ResponseWriter, _ *http.Request) {
	if err := s.DB.PurgeAllData(); err != nil {
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "purged"})
}
