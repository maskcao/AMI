package api

import (
	"net/http"
)

func (s *Server) getIntelMarkdown(w http.ResponseWriter, r *http.Request) {
	id := pathID(r)
	mdPath, _ := s.DB.ParsedMdPath(id)
	if mdPath == "" {
		writeErr(w, http.StatusNotFound, "no markdown")
		return
	}
	b, err := s.Files.ReadParsedText(s.DB.DataDir(), mdPath)
	if err != nil || len(b) == 0 {
		writeErr(w, http.StatusNotFound, "no markdown")
		return
	}
	w.Header().Set("Content-Type", "text/markdown; charset=utf-8")
	_, _ = w.Write(b)
}
